package service

import (
	"reflect"
	"sync"

	"github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/rpcserver"
)

func NewServiceClientPools(s *Service, rpcs *rpcserver.RpcServer) rpcserver.RpcPool {
	return &ServiceClientPools{pls: sync.Pool{
		New: func() interface{} {
			c := &ServiceClient{RpcClient: rpcserver.NewRpcClient(rpcs),_service: s}
			c.Register(reflect.TypeOf(&protocols.RegisterReq{}), c.onRegisterReq)
			return c
		},
	}}
}

type ServiceClientPools struct {
	pls sync.Pool
}

func (drp *ServiceClientPools) Get() network.Client {
	return drp.pls.Get().(network.Client)
}

func (drp *ServiceClientPools) Put(s *rpcserver.RpcClient) {
	// s.clientID = nil
	// s.recvice.Reset()
	drp.pls.Put(s)
}
