package service

import (
	"reflect"
	"sync"

	"github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/rpcserver"
)

func NewServiceClientPool(s *Service,register func(key string, clientId *network.ClientID),unregister func(key string, clientId *network.ClientID)) rpcserver.RpcPool {
	return &ServiceClientPool{pls: sync.Pool{
		New: func() interface{} {
			c := &ServiceClient{RpcClient: rpcserver.NewRpcClient(s.RpcServer)}
			c.register =register
			c.unregister = unregister
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
	s.Destory()
	drp.pls.Put(s)
}
