package client

import (

	"github.com/yamakiller/velcro-go/example/monopoly/generate/protocols"
	"github.com/yamakiller/velcro-go/rpc/server"
)

func (lc *LoginClient) onRegisterAccountRequest(ctx *server.RpcClientContext) interface{} {
 	request :=	ctx.Message().(*protocols.RegisterAccountRequest)
	ctx.Context().Debug("account %s  pass :%s",request.Account,request.Pass )
	return &protocols.RegisterAccountResponse{Res: 1}
}

func (lc *LoginClient) onSigninRequest(ctx *server.RpcClientContext) interface{}  {
	request :=	ctx.Message().(*protocols.SigninRequest)
	ctx.Context().Debug("account %s  pass :%s",request.Account,request.Pass )
	return nil
}

// Logout 退出登录
func (lc *LoginClient) onSignoutRequest(ctx *server.RpcClientContext) interface{} {
	return nil
}
