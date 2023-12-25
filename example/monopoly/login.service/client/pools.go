package client

import (
	"reflect"
	"sync"

	"github.com/yamakiller/velcro-go/cluster/service"
	"github.com/yamakiller/velcro-go/example/monopoly/generate/protocols"
	"github.com/yamakiller/velcro-go/rpc/server"
)

func NewLoginServiceClientPool(s *LoginService) server.RpcPool {
	return &LoginServiceClientPool{pls: sync.Pool{
		New: func() interface{} {
			c := &LoginClient{}
			c.ServiceClient = service.NewServiceClient(s.Service,s.RegisetrGroup,s.UnregisterGroup)
			c.Register(reflect.TypeOf(&protocols.RegisterAccountRequest{}), c.onRegisterAccountRequest)
			c.Register(reflect.TypeOf(&protocols.SigninRequest{}), c.onSigninRequest)
			c.Register(reflect.TypeOf(&protocols.SignoutRequest{}), c.onSignoutRequest)
			// c.Register(reflect.TypeOf(&protocols.RegisterRequest{}), c.onRegister)
			return c
		},
	}}
}

type LoginServiceClientPool struct {
	pls sync.Pool
}

func (drp *LoginServiceClientPool) Get() server.RpcClient {
	return drp.pls.Get().(server.RpcClient)
}

func (drp *LoginServiceClientPool) Put(s server.RpcClient) {
	s.Destory()
	drp.pls.Put(s)
}