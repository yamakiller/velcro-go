package service

import (
	"github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/rpcserver"
)

type ServiceClient struct {
	rpcserver.RpcClient
	VAddr string
	Tag   string
}

func (sc *ServiceClient) onRegister(ctx *rpcserver.RpcClientContext) interface{} {
	//requst := ctx.Message().(*protocols.RegisterRequest)
	//sc.VAddr = requst.Vaddr
	//sc.Tag = requst.Tag

	return &protocols.RegisterResponse{}
}

func (sc *ServiceClient) Closed(ctx network.Context) {

	sc.RpcClient.Closed(ctx)
}
