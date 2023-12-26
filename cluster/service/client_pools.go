package service

import (
	"sync"

	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/server"
)

func NewServiceClientPool(s *Service,
	reg func(key string, clientId *network.ClientID),
	unreg func(key string, clientId *network.ClientID)) server.RpcPool {
	return &ServiceClientPool{pls: sync.Pool{
		New: func() interface{} {
			c:= NewServiceClient(s, reg, unreg)		
			return c
		},
	}}
}

type ServiceClientPool struct {
	pls sync.Pool
}

func (drp *ServiceClientPool) Get() server.RpcClient {
	return drp.pls.Get().(server.RpcClient)
}

func (drp *ServiceClientPool) Put(s server.RpcClient) {
	s.Destory()
	drp.pls.Put(s)
}
