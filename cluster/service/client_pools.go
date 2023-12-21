package service

import (
	"reflect"
	"sync"

	"github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/rpc/rpcserver"
)

func NewServiceClientPool(s *Service) rpcserver.RpcPool {
	return &ServiceClientPool{pls: sync.Pool{
		New: func() interface{} {
			c := &ServiceClient{RpcClient: rpcserver.NewRpcClient(s.RpcServer)}
			c.Register(reflect.TypeOf(&protocols.RegisterRequest{}), c.onRegister)
			return c
		},
	}}
}

type ServiceClientPool struct {
	pls sync.Pool
}

func (drp *ServiceClientPool) Get() rpcserver.RpcClient {
	return drp.pls.Get().(rpcserver.RpcClient)
}

func (drp *ServiceClientPool) Put(s rpcserver.RpcClient) {
	// s.clientID = nil
	// s.recvice.Reset()
	drp.pls.Put(s)
}
