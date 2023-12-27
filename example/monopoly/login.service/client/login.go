package client

import (
	"time"

	clusterprotocols "github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/example/monopoly/generate/protocols"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/server"
)

func (lc *LoginClient) onSignIn(ctx *server.RpcClientContext, msg interface{}, sender *network.ClientID) interface{} {
	request := msg.(*protocols.SigninRequest)

	player, err := accounts.SignIn(ctx.Background(), request.Token)
	if err != nil {

		if err = lc.PostMessage(ctx.Context(), &clusterprotocols.Closing{
			ClientID: sender,
		}); err != nil {
			ctx.Context().Error("post closing %s fail", sender.ToString())
		}

		ctx.Context().Error("onSignin error %s", err.Error())
		return nil
	}

	if err := rds.RegisterPlayer(ctx.Background(), player.UID, sender, player, 10*time.Minute); err != nil {
		if err = lc.PostMessage(ctx.Context(), &clusterprotocols.Closing{
			ClientID: sender,
		}); err != nil {
			ctx.Context().Error("post closing %s fail", sender.ToString())
		}

		return nil
	}
	res := &protocols.SigninResponse{}
	res.Name = player.UID
	res.Externs = make(map[string]string)
	for k, v := range player.Externs {
		res.Externs[k] = v
	}
	return res
}

// Logout 退出登录
func (lc *LoginClient) onSignOut(ctx *server.RpcClientContext, msg interface{}, sender *network.ClientID) interface{} {

	isAuth, err := rds.IsAuth(ctx.Background(), sender)
	if err != nil {
		ctx.Context().Error("client %s unauthorized or closed", sender.ToString())
		lc.PostClientClose(ctx, sender)
		return &protocols.SignoutResponse{Res: -1}
	}

	if !isAuth {
		lc.PostClientClose(ctx, sender)
		ctx.Context().Error("client %s unauthorized", sender.ToString())
		return &protocols.SignoutResponse{Res: -2}
	}

	// 需要加入token
	if err := accounts.SignOut(); err != nil {
		ctx.Context().Error("client %s signout fail %s", sender.ToString(), err.Error())
		return &protocols.SignoutResponse{Res: -3}
	}

	if err := rds.UnRegisterPlayer(ctx.Background(), sender); err != nil {
		ctx.Context().Error("client %s SignOut fail %s", sender.ToString(), err.Error())
	}

	return &protocols.SignoutResponse{Res: 0}
}
