package client

import (
	"github.com/yamakiller/velcro-go/example/monopoly/generate/protocols"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	"github.com/yamakiller/velcro-go/rpc/server"
)

func (lc *LoginClient) onRegisterAccountRequest(ctx *server.RpcClientContext) interface{} {
 	request :=	ctx.Message().(*protocols.RegisterAccountRequest)
	ctx.Context().Debug("account %s  pass :%s",request.Account,request.Pass )
	return &protocols.RegisterAccountResponse{Res: 1}
}

func (lc *LoginClient) onSigninRequest(ctx *server.RpcClientContext) interface{}  {
	request :=	ctx.Message().(*protocols.SigninRequest)
	
	player,err :=accounts.SignIn(request.Token)
	if err !=  nil{
		return nil
	}
	lc.name = player.Name
	if err := rds.PushUser(lc.name,player);err != nil{
		return nil
	}
	res := &protocols.SigninResponse{}
	res.Name = lc.name
	res.Externs = make(map[string]string)
	for k,v :=range player.Externs{
		res.Externs[k] =v
	}
	return res
}

// Logout 退出登录
func (lc *LoginClient) onSignoutRequest(ctx *server.RpcClientContext) interface{} {

	
	if err :=accounts.SignOut();err !=  nil{
		return nil
	}

	if err := rds.RemUser(lc.name);err != nil{
		return nil
	}

	return &protocols.SignoutResponse{Res: 0}
}
