package apps

import (
	"time"

	clusterprotocols "github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/example/monopoly/generate/protocols"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils"
	"google.golang.org/protobuf/proto"
)

func newActor(conn *serve.ServantClientConn) serve.ServantClientActor {
	actor := &LoginActor{
		ancestor: conn.Servant,
	}

	conn.Register(&protocols.SigninRequest{}, actor.onSignIn)
	conn.Register(&protocols.SignoutRequest{}, actor.onSignIn)
	conn.Register(&clusterprotocols.Closed{}, actor.onOtherClosed)

	return actor
}

type LoginActor struct {
	ancestor *serve.Servant
}

func (actor *LoginActor) onSignIn(ctx *serve.ServantClientContext) (proto.Message, error) {
	request := ctx.Message.(*protocols.SigninRequest)
	utils.AssertEmpty(request, "onSignIn message not protocols.SigninRequest")
	utils.AssertEmpty(ctx.Sender, "onSignIn sender is null")

	player, err := accounts.SignIn(ctx.Background, request.Token)
	if err != nil {
		actor.closeClient(ctx, ctx.Sender)
		ctx.Context.Debug("onSignin error %s", err.Error())
		return nil, err
	}

	if err := rds.RegisterPlayer(ctx.Background, player.UID, ctx.Sender, player, 10*time.Minute); err != nil {
		actor.closeClient(ctx, ctx.Sender)
		return nil, err
	}

	resp := &protocols.SigninResponse{}
	resp.Name = player.UID
	resp.Externs = make(map[string]string)
	for k, v := range player.Externs {
		resp.Externs[k] = v
	}
	return resp, nil
}

func (actor *LoginActor) onSignOut(ctx *serve.ServantClientContext) (proto.Message, error) {
	return nil, nil
}

func (actor *LoginActor) onOtherClosed(ctx *serve.ServantClientContext) {

}

// Closed 自己被关闭
func (actor *LoginActor) Closed(ctx *serve.ServantClientContext) {

}

// closeClient 关闭某个客户端
func (actor *LoginActor) closeClient(ctx *serve.ServantClientContext, client *network.ClientID) {
	if closError := actor.ancestor.PostMessage(client.Address, &clusterprotocols.Closing{
		ClientID: client,
	}); closError != nil {
		ctx.Context.Error("post closing %s fail", client.ToString())
	}
}
