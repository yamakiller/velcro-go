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
	tag   string

	register   func(key string, clientId *network.ClientID)
	unregister func(key string, clientId *network.ClientID)
}

func (sc *ServiceClient) onRegister(ctx *rpcserver.RpcClientContext) interface{} {
	requst := ctx.Message().(*protocols.RegisterRequest)
	utils.AssertEmpty(requst, "Service Client onRegister message error is nil")
	sc.vaddr = requst.Vaddr
	sc.tag = requst.Tag
	sc.register(sc.toKey(), ctx.Context().Self())

	return &protocols.RegisterResponse{}
}

func (sc *ServiceClient) Closed(ctx network.Context) {
	if sc.vaddr != "" && sc.tag != "" {
		sc.unregister(sc.toKey())
	}

	sc.RpcClient.Closed(ctx)
}

func (sc *ServiceClient) toKey() string {
	return sc.tag + sc.vaddr
}
