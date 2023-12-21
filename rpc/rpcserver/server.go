package rpcserver

import (
	"sync"

	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/rpcmessage"
)

func New(options ...ConnConfigOption) *RpcServer {
	config := Configure(options...)
	s := &RpcServer{clients: make(map[network.CIDKEY]RpcClient)}
	s.NetworkSystem = network.NewTCPServerNetworkSystem(network.WithProducer(s.newClient))

	if config.Pool == nil {
		config.Pool = NewDefaultRpcPool(s)
	}
	s.clientPool = config.Pool
	s.MarshalMessage = config.MarshalMessage
	s.MarshalPing = config.MarshalPing
	s.MarshalResponse = config.MarshalResponse
	s.UnMarshal = config.UnMarshal

	return s
}

type RpcServer struct {
	*network.NetworkSystem
	sync.Mutex
	clients    map[network.CIDKEY]RpcClient
	clientPool RpcPool

	MarshalResponse rpcmessage.MarshalResponseFunc
	MarshalMessage  rpcmessage.MarshalMessageFunc
	MarshalPing     rpcmessage.MarshalPingFunc

	UnMarshal rpcmessage.UnMarshalFunc
}

func (s *RpcServer) WithPool(pool RpcPool) {
	s.clientPool = pool
}

func (s *RpcServer) Open(addr string) error {
	return s.NetworkSystem.Open(addr)
}

func (s *RpcServer) Shutdown() {
	s.Lock()
	for _, c := range s.clients {
		c.ClientID().UserClose()
	}
	s.Unlock()

	s.NetworkSystem.Shutdown()
}

func (s *RpcServer) Register(cid *network.ClientID, rc RpcClient) {
	s.Lock()
	defer s.Unlock()

	_, ok := s.clients[network.Key(cid)]
	if ok {
		panic("register linker error: ClientID Repeat")
	}

	rc.referenceIncrement()
	s.clients[network.Key(cid)] = rc
}

func (s *RpcServer) UnRegister(cid *network.ClientID) {

	s.Lock()
	l, ok := s.clients[network.Key(cid)]
	if !ok {
		s.Unlock()
		return
	}

	ref := l.referenceDecrement()
	delete(s.clients, network.Key(cid))

	s.Unlock()

	if ref == 0 {
		s.free(l)
	}
}

func (s *RpcServer) GetClient(cid *network.ClientID) RpcClient {
	s.Lock()
	defer s.Unlock()

	l, ok := s.clients[network.Key(cid)]
	if !ok {
		return nil
	}

	l.referenceIncrement()
	return l
}

func (s *RpcServer) ReleaseClient(client RpcClient) {
	s.Lock()
	ref := client.referenceDecrement()
	s.Unlock()

	if ref == 0 {
		s.free(client)
	}
}

func (s *RpcServer) newClient(system *network.NetworkSystem) network.Client {
	return s.clientPool.Get()
}

func (s *RpcServer) free(rc RpcClient) {
	s.clientPool.Put(rc)
}
