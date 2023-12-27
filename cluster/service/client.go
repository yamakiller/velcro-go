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

func NewServiceClient(s *Service, register FRegister, unregister FUnRegister) *ServiceClient {
	c := &ServiceClient{
		RpcClient:      server.NewRpcClientConn(s.RpcServer),
		forwardMethods: make(map[interface{}]func(*server.RpcClientContext, interface{}, *network.ClientID) interface{}),
		register:       register,
		unregister:     unregister,
	}

	c.Register(&protocols.RegisterRequest{}, c.onRegister)
	c.Register(&protocols.Forward{}, c.onForward)

	return c
}

type ServiceClient struct {
	server.RpcClient

	vaddr          string
	forwardMethods map[interface{}]func(*server.RpcClientContext, interface{}, *network.ClientID) interface{}

	register   func(key string, clientId *network.ClientID)
	unregister func(key string, clientId *network.ClientID)
}

func (sc *ServiceClient) RegisterForward(key interface{}, f func(*server.RpcClientContext, interface{}, *network.ClientID) interface{}) {
	sc.forwardMethods[reflect.TypeOf(key)] = f
}

// onRegister 连接者注册到服务中.
// 主要用于识别标记
func (sc *ServiceClient) onRegister(ctx *server.RpcClientContext) interface{} {
	requst := ctx.Message().(*protocols.RegisterRequest)
	utils.AssertEmpty(requst, "Service Client onRegister message error is nil")
	sc.vaddr = requst.Vaddr
	sc.register(sc.vaddr, ctx.Context().Self())

	return &protocols.RegisterResponse{}
}

// onForward 接受其它服务托送来的数据
func (sc *ServiceClient) onForward(ctx *server.RpcClientContext) interface{} {
	requst := ctx.Message().(*protocols.Forward)
	utils.AssertEmpty(requst, "Service Client onForward message error is nil")

	msg, err := requst.Msg.UnmarshalNew()
	if err != nil {
		sc.PostClientClose(ctx, requst.Sender)
		if ctx.Background() != nil {
			return sc.NewError("Serive System error", "decoding failed")
		} else {
			return nil
		}
	}

	f, ok := sc.forwardMethods[reflect.TypeOf(msg)]
	if !ok {

		sc.PostClientClose(ctx, requst.Sender)
		if ctx.Background() != nil {
			return sc.NewError("Serive System error", "unfound specify method")
		} else {
			return nil
		}
	}

	return f(ctx, msg, requst.Sender)
}

func (sc *ServiceClient) Closed(ctx network.Context) {

	if sc.vaddr != "" {
		sc.unregister(sc.vaddr, ctx.Self())
	}

	sc.RpcClient.Closed(ctx)
}

func (sc *ServiceClient) PostClientClose(ctx *server.RpcClientContext, clientId *network.ClientID) {
	if err := sc.PostMessage(ctx.Context(), &protocols.Closing{
		ClientID: clientId,
	}); err != nil {
		ctx.Context().Debug("post closing %s fail", clientId.ToString())
	}
}

func (sc *ServiceClient) NewError(name, msg string) *protocols.Error {
	return &protocols.Error{
		ID:   0,
		Name: name,
		Err:  msg,
	}
}
