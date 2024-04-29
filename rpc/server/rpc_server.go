package server

import (
	"sync"

	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/circbuf"
)

func New(options ...ConnConfigOption) *RpcServer {
	config := Configure(options...)
	s := &RpcServer{connects: make(map[network.CIDKEY]IRpcConnector)}
	s.NetworkSystem = network.NewTCPServerNetworkSystem(network.WithProducer(s.connectAlloc))

	if config.Allocator == nil {
		config.Allocator = &defaultRpcAllocator{
			m: sync.Pool{
				New: func() interface{} {
					return &RpcConnector{
						recvice:   circbuf.NewLinkBuffer(32),
						reference: 0,
					}
				},
			},
		}
	}

	s.connectAllocator = config.Allocator

	return s
}

type RpcServer struct {
	*network.NetworkSystem

	connects         map[network.CIDKEY]IRpcConnector
	connectAllocator IRpcAllocator
	connectMutex     sync.Mutex
}

func (s *RpcServer) Open(addr string) error {
	return s.NetworkSystem.Open(addr)
}

func (s *RpcServer) Shutdown() {
	s.connectMutex.Lock()
	for _, c := range s.connects {
		c.Destory() //GetID().UserClose()
	}
	s.connectMutex.Unlock()

	s.NetworkSystem.Shutdown()
}

func (s *RpcServer) Register(cid *network.ClientID, rc IRpcConnector) {
	s.connectMutex.Lock()
	defer s.connectMutex.Unlock()

	_, ok := s.connects[network.Key(cid)]
	if ok {
		panic("register linker error: ClientID Repeat")
	}

	rc.referenceIncrement()
	s.connects[network.Key(cid)] = rc
}

func (s *RpcServer) UnRegister(cid *network.ClientID) {

	s.connectMutex.Lock()
	l, ok := s.connects[network.Key(cid)]
	if !ok {
		s.connectMutex.Unlock()
		return
	}

	ref := l.referenceDecrement()
	delete(s.connects, network.Key(cid))

	s.connectMutex.Unlock()

	if ref == 0 {
		s.connectFree(l)
	}
}

func (s *RpcServer) PickUpConnector(cid *network.ClientID) IRpcConnector {
	s.connectMutex.Lock()
	defer s.connectMutex.Unlock()

	l, ok := s.connects[network.Key(cid)]
	if !ok {
		return nil
	}

	l.referenceIncrement()
	return l
}

func (s *RpcServer) ReleaseConnector(conn IRpcConnector) {
	s.connectMutex.Lock()
	ref := conn.referenceDecrement()
	s.connectMutex.Unlock()

	if ref == 0 {
		s.connectFree(conn)
	}
}

func (s *RpcServer) connectAlloc(system *network.NetworkSystem) network.Client {
	return s.connectAllocator.Get()
}

func (s *RpcServer) connectFree(rc IRpcConnector) {
	s.connectAllocator.Put(rc)
}
