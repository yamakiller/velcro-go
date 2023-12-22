package service

import (
	"github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/rpcserver"
	"github.com/yamakiller/velcro-go/utils"
)

type ServiceClient struct {
	rpcserver.RpcClient

	vaddr string

	register   func(key string, clientId *network.ClientID)
	unregister func(key string, clientId *network.ClientID)
}

func (sc *ServiceClient) onRegister(ctx *rpcserver.RpcClientContext) interface{} {
	requst := ctx.Message().(*protocols.RegisterRequest)
	utils.AssertEmpty(requst, "Service Client onRegister message error is nil")
	sc.vaddr = requst.Vaddr
	sc.register(sc.vaddr, ctx.Context().Self())

	return &protocols.RegisterResponse{}
}

func (sc *ServiceClient) Closed(ctx network.Context) {

	if sc.vaddr != "" {
		sc.unregister(sc.vaddr, ctx.Self())
	}

	sc.RpcClient.Closed(ctx)
}
