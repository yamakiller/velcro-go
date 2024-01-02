package apps

import (
	"errors"

	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils"
	"google.golang.org/protobuf/proto"
)

const (
	defaultRequestTimeout = 1000
)

type LoginActor struct {
	ancestor *serve.Servant
}

func (actor *LoginActor) onSignIn(ctx *serve.ServantClientContext) (proto.Message, error) {
	request := ctx.Message.(*mpubs.SignIn)
	utils.AssertEmpty(request, "onSignIn message not protocols.SignIn")
	utils.AssertEmpty(ctx.Sender, "onSignIn sender is null")

	player, err := accounts.SignIn(ctx.Background, request.Token)
	if err != nil {
		actor.submitRequestCloseClient(ctx, ctx.Sender)
		ctx.Context.Debug("onSignin error %s", err.Error())
		return nil, err
	}

	if err := rds.RegisterPlayer(ctx.Background, ctx.Sender,
		player.UID, player.DisplayName, player); err != nil {
		actor.submitRequestCloseClient(ctx, ctx.Sender)
		return nil, err
	}

	resp := &mpubs.SignInResp{}
	resp.Uid = player.UID
	resp.DisplayName = player.DisplayName
	resp.Externs = make(map[string]string)
	for k, v := range player.Externs {
		resp.Externs[k] = v
	}
	return resp, nil
}

func (actor *LoginActor) onSignOut(ctx *serve.ServantClientContext) (proto.Message, error) {
	request := ctx.Message.(*mpubs.SignOut)
	utils.AssertEmpty(request, "onSignOut message not pubs.SignOut")
	utils.AssertEmpty(ctx.Sender, "onSignOut sender is null")

	accounts.SignOut(request.Token)

	var (
		uid     string
		err     error
		results map[string]string
	)

	uid, err = rds.GetPlayerUID(ctx.Background, ctx.Sender)
	if err != nil {
		return nil, err
	}

	if results, err = rds.UnRegisterPlayer(ctx.Background, ctx.Sender, uid); err != nil {
		actor.submitRequestCloseClient(ctx, ctx.Sender)
		return nil, err
	}

	if results == nil {
		return &mpubs.SignOutResp{
			Result: 0,
		}, nil
	}

	if battleSpaceId, ok := results[rdsconst.PlayerMapClientBattleSpaceId]; ok && battleSpaceId != "" {
		actor.submitRequest(ctx, &mprvs.RequestExitBattleSpace{BattleSpaceID: battleSpaceId, UID: uid})
		// TODO: 是否需要判断执行失败
	}

	return &mpubs.SignOutResp{
		Result: 0,
	}, nil
}

func (actor *LoginActor) submitRequestCloseClient(ctx *serve.ServantClientContext, clientId *network.ClientID) {
	actor.submitRequest(ctx, &prvs.RequestGatewayCloseClient{Target: clientId})
}

func (actor *LoginActor) submitRequest(ctx *serve.ServantClientContext, request proto.Message) (proto.Message, error) {
	r := actor.ancestor.FindRouter(request)
	if r == nil {
		if ctx != nil {
			ctx.Context.Error("%s unfound router", proto.MessageName(request))
		}

		return nil, errors.New("unfound router")
	}

	result, err := r.Proxy.RequestMessage(request, defaultRequestTimeout)
	if err != nil {
		if ctx != nil {
			ctx.Context.Error("%s fail error %s", proto.MessageName(request), err.Error())
		}
	}

	return result, err
}

func (actor *LoginActor) Closed(ctx *serve.ServantClientContext) {

}
