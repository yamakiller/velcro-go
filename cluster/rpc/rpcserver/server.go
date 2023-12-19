package rpcserver

import (
	"github.com/yamakiller/velcro-go/cluster/rpc/rpcmessage"
	"github.com/yamakiller/velcro-go/network"
)

type RpcServer struct {
	network.NetworkSystem
	group      *ClientGroup
	clientPool RpcPool

	MarshalResponse rpcmessage.MarshalResponseFunc
	MarshalMessage  rpcmessage.MarshalMessageFunc
	MarshalPing     rpcmessage.MarshalPingFunc

	UnMarshal rpcmessage.UnMarshalFunc
}

func (s *RpcServer) Register(cid *network.ClientID, rc *RpcClient) {
	s.group.mu.Lock()
	defer s.group.mu.Unlock()

	_, ok := s.group.clients[network.Key(cid)]
	if ok {
		panic("register linker error: ClientID Repeat")
	}

	rc.referenceIncrement()
	s.group.clients[network.Key(cid)] = rc
}

func (s *RpcServer) UnRegister(cid *network.ClientID) {

	s.group.mu.Lock()
	l, ok := s.group.clients[network.Key(cid)]
	if !ok {
		s.group.mu.Unlock()
		return
	}

	ref := l.referenceDecrement()
	delete(s.group.clients, network.Key(cid))

	s.group.mu.Unlock()

	if ref == 0 {
		s.free(l)
	}
}

func (s *RpcServer) GetClient(cid *network.ClientID) *RpcClient {
	s.group.mu.Lock()
	defer s.group.mu.Unlock()

	l, ok := s.group.clients[network.Key(cid)]
	if !ok {
		return nil
	}

	l.referenceIncrement()
	return l
}

func (s *RpcServer) ReleaseClient(client *RpcClient) {
	s.group.mu.Lock()
	ref := client.referenceDecrement()
	s.group.mu.Unlock()

	if ref == 0 {
		s.free(client)
	}
}

func (s *RpcServer) newClient(system *network.NetworkSystem) network.Client {
	return s.clientPool.Get()
}

func (s *RpcServer) free(rc *RpcClient) {
	s.clientPool.Put(rc)
}
