package service

import (
	"context"

	"github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/rpcserver"
)

type ServiceClient struct {
	*rpcserver.RpcClient
	_service *Service
	_addr string
	_tag string
}

func (sc *ServiceClient) onRegisterReq(timeout context.Context, ctx network.Context, message interface{}) interface{} {
	requst := message.(*protocols.RegisterReq)
	sc._addr = requst.Vaddr
	sc._tag = requst.Tag
	sc._service.Register(ctx.Self(), requst.Vaddr,requst.Tag)
	response := &protocols.RegisterResp{}
	return response
}