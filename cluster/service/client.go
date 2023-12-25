package service

import (
	"reflect"

	"github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/server"
	"github.com/yamakiller/velcro-go/utils"
)
type FRegister func(key string, clientId *network.ClientID)
type FUnRegister func(key string, clientId *network.ClientID)

func NewServiceClient(s *Service,register FRegister, unregister FUnRegister) *ServiceClient {
	c := &ServiceClient{
		RpcClient: server.NewRpcClientConn(s.RpcServer),
		register:   register,
		unregister: unregister,
	}
	c.Register(reflect.TypeOf(&protocols.RegisterRequest{}), c.onRegister)
	// c.Register(reflect.TypeOf(&protocols.ClientRequestMessage{}), c.onClientRequestMessage)
	return c
}


type ServiceClient struct {
	server.RpcClient

	vaddr string

	register   func(key string, clientId *network.ClientID)
	unregister func(key string, clientId *network.ClientID)
}

func (sc *ServiceClient) onRegister(ctx *server.RpcClientContext) interface{} {
	requst := ctx.Message().(*protocols.RegisterRequest)
	utils.AssertEmpty(requst, "Service Client onRegister message error is nil")
	sc.vaddr = requst.Vaddr
	sc.register(sc.vaddr, ctx.Context().Self())

	return &protocols.RegisterResponse{}
}
// func (sc *ServiceClient)onClientRequestMessage(ctx *server.RpcClientContext) interface{} {
// 	request := ctx.Message().(*protocols.ClientRequestMessage)

// }
func (sc *ServiceClient) Closed(ctx network.Context) {

	if sc.vaddr != "" {
		sc.unregister(sc.vaddr, ctx.Self())
	}

	sc.RpcClient.Closed(ctx)
}
