package server

import (
	"sync"

	"github.com/yamakiller/velcro-go/network"
	rpcmessage "github.com/yamakiller/velcro-go/rpc/messages"
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

	clients    map[network.CIDKEY]RpcClient
	clientPool RpcPool
	climu      sync.Mutex

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
	s.climu.Lock()
	for _, c := range s.clients {
		c.ClientID().UserClose()
	}
	s.climu.Unlock()

	s.NetworkSystem.Shutdown()
}

func (s *RpcServer) Register(cid *network.ClientID, rc RpcClient) {
	s.climu.Lock()
	defer s.climu.Unlock()

	_, ok := s.clients[network.Key(cid)]
	if ok {
		panic("register linker error: ClientID Repeat")
	}

	rc.referenceIncrement()
	s.clients[network.Key(cid)] = rc
}

func (s *RpcServer) UnRegister(cid *network.ClientID) {

	s.climu.Lock()
	l, ok := s.clients[network.Key(cid)]
	if !ok {
		s.climu.Unlock()
		return
	}

	ref := l.referenceDecrement()
	delete(s.clients, network.Key(cid))

	s.climu.Unlock()

	if ref == 0 {
		s.free(l)
	}
}

func (s *RpcServer) GetClient(cid *network.ClientID) RpcClient {
	s.climu.Lock()
	defer s.climu.Unlock()

	l, ok := s.clients[network.Key(cid)]
	if !ok {
		return nil
	}

	l.referenceIncrement()
	return l
}

func (s *RpcServer) ReleaseClient(client RpcClient) {
	s.climu.Lock()
	ref := client.referenceDecrement()
	s.climu.Unlock()

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
