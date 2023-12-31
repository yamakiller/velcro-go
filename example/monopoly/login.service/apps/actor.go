package apps

import (
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	"github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils"
	"google.golang.org/protobuf/proto"
)

/*
import (

	"context"
	"time"

	clusterprotocols "github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/protocols"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils"
	"google.golang.org/protobuf/proto"

)

	func newActor(conn *serve.ServantClientConn) serve.ServantClientActor {
		actor := &LoginActor{
			ancestor: conn.Servant,
		}

		conn.Register(&protocols.SignIn{}, actor.onSignIn)
		conn.Register(&protocols.SignOut{}, actor.onSignIn)
		conn.Register(&clusterprotocols.Closed{}, actor.onOtherClosed)

		return actor
	}

	type LoginActor struct {
		ancestor *serve.Servant
	}

	func (actor *LoginActor) onSignIn(ctx *serve.ServantClientContext) (proto.Message, error) {
		request := ctx.Message.(*protocols.SignIn)
		utils.AssertEmpty(request, "onSignIn message not protocols.SignIn")
		utils.AssertEmpty(ctx.Sender, "onSignIn sender is null")

		player, err := accounts.SignIn(ctx.Background, request.Token)
		if err != nil {
			actor.closeClient(ctx, ctx.Sender)
			ctx.Context.Debug("onSignin error %s", err.Error())
			return nil, err
		}

		if err := rds.RegisterPlayer(ctx.Background, ctx.Sender,
			player.UID, player.DisplayName, player); err != nil {
			actor.closeClient(ctx, ctx.Sender)
			return nil, err
		}

		resp := &protocols.SignInResp{}
		resp.Uid = player.UID
		resp.DisplayName = player.DisplayName
		resp.Externs = make(map[string]string)
		for k, v := range player.Externs {
			resp.Externs[k] = v
		}
		return resp, nil
	}

	func (actor *LoginActor) onSignOut(ctx *serve.ServantClientContext) (proto.Message, error) {
		request := ctx.Message.(*protocols.SignOut)
		utils.AssertEmpty(request, "onSignOut message not protocols.SignOut")
		utils.AssertEmpty(ctx.Sender, "onSignOut sender is null")

		accounts.SignOut(request.Token)

		var (
			err     error
			results map[string]string
		)
		if results, err = rds.UnRegisterPlayer(ctx.Background, ctx.Sender); err != nil {
			actor.closeClient(ctx, ctx.Sender)
			return nil, err
		}

		if results == nil {
			return &protocols.SignOutResp{
				Result: 0,
			}, nil
		}

		if battleSpaceId, ok := results[rdsconst.PlayerMapClientBattleSpaceId]; ok && battleSpaceId != "" {
			// TODO: 通知房间服务器, 某用户已退出`	`
		}

		return &protocols.SignOutResp{
			Result: 0,
		}, nil
	}

	func (actor *LoginActor) onOtherClosed(ctx *serve.ServantClientContext) {
		message := ctx.Message.(*clusterprotocols.Closed)
		utils.AssertEmpty(message, "onOtherClosed message not clusterprotocols.Closed")
		background, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		results, err := rds.UnRegisterPlayer(background, message.ClientID)
		if err != nil {
			ctx.Context.Error("onOtherClosed, UnRegisterPlayer %s", err.Error())
			return
		}

		if results == nil {
			return
		}

		if battleSpaceId, ok := results[rdsconst.PlayerMapClientBattleSpaceId]; ok && battleSpaceId != "" {
			// TODO: 通知房间服务器, 某用户已退出
		}
	}

// Closed 自己被关闭
func (actor *LoginActor) Closed(ctx *serve.ServantClientContext) {

}

// closeClient 关闭某个客户端

	func (actor *LoginActor) closeClient(ctx *serve.ServantClientContext, client *network.ClientID) {
		if closError := actor.ancestor.ReplyMessage(client.Address, &clusterprotocols.Closing{
			ClientID: client,
		}); closError != nil {
			ctx.Context.Error("post closing %s fail", client.ToString())
		}
	}
*/
const (
	defaultRequestTimeout = 1000
)

type LoginActor struct {
	ancestor *serve.Servant
}

func (actor *LoginActor) onSignIn(ctx *serve.ServantClientContext) (proto.Message, error) {
	request := ctx.Message.(*pubs.SignIn)
	utils.AssertEmpty(request, "onSignIn message not protocols.SignIn")
	utils.AssertEmpty(ctx.Sender, "onSignIn sender is null")

	player, err := accounts.SignIn(ctx.Background, request.Token)
	if err != nil {
		actor.RequestCloseClient(ctx, ctx.Sender)
		ctx.Context.Debug("onSignin error %s", err.Error())
		return nil, err
	}

	if err := rds.RegisterPlayer(ctx.Background, ctx.Sender,
		player.UID, player.DisplayName, player); err != nil {
		actor.RequestCloseClient(ctx, ctx.Sender)
		return nil, err
	}

	resp := &pubs.SignInResp{}
	resp.Uid = player.UID
	resp.DisplayName = player.DisplayName
	resp.Externs = make(map[string]string)
	for k, v := range player.Externs {
		resp.Externs[k] = v
	}
	return resp, nil
}

func (actor *LoginActor) onSignOut(ctx *serve.ServantClientContext) (proto.Message, error) {
	request := ctx.Message.(*pubs.SignOut)
	utils.AssertEmpty(request, "onSignOut message not pubs.SignOut")
	utils.AssertEmpty(ctx.Sender, "onSignOut sender is null")

	accounts.SignOut(request.Token)

	var (
		err     error
		results map[string]string
	)
	if results, err = rds.UnRegisterPlayer(ctx.Background, ctx.Sender); err != nil {
		actor.RequestCloseClient(ctx, ctx.Sender)
		return nil, err
	}

	if results == nil {
		return &pubs.SignOutResp{
			Result: 0,
		}, nil
	}

	if battleSpaceId, ok := results[rdsconst.PlayerMapClientBattleSpaceId]; ok && battleSpaceId != "" {
		// TODO: 通知房间服务器, 某用户已退出`	`
	}

	return &pubs.SignOutResp{
		Result: 0,
	}, nil
}

func (actor *LoginActor) RequestCloseClient(ctx *serve.ServantClientContext, clientId *network.ClientID) {
	request := &prvs.RequestGatewayCloseClient{Target: clientId}
	r := actor.ancestor.FindRouter(request)
	if r == nil {
		if ctx != nil {
			ctx.Context.Error("RequestCloseClient unfound router")
		}

		return
	}

	if _, err := r.Proxy.RequestMessage(request, defaultRequestTimeout); err != nil {
		if ctx != nil {
			ctx.Context.Error("RequestCloseClient fail error %s", err.Error())
		}
		return
	}
}

func (actor *LoginActor) Closed(ctx *serve.ServantClientContext) {

}
