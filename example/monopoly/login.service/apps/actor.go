package apps

import (
	"context"
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
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
)

const (
	defaultRequestTimeout = 2000
)

type LoginActor struct {
	ancestor *serve.Servant
}

func (actor *LoginActor) onSignIn(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.SignIn)
	sender :=serve.GetServantClientInfo(ctx).Sender()
	utils.AssertEmpty(request, "onSignIn message not protocols.SignIn")
	utils.AssertEmpty(sender, "onSignIn sender is null")

	player, err := accounts.SignIn(ctx, request.Token)
	if err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.ContextDebugf(ctx, "onSignin error %s", err.Error())
		return nil, err
	}

	if err := rds.RegisterPlayer(ctx, sender,
		player.UID, player.DisplayName, player); err != nil {
		actor.submitRequestCloseClient(ctx, sender)
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

func (actor *LoginActor) onSignOut(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.SignOut)
	sender :=serve.GetServantClientInfo(ctx).Sender()
	utils.AssertEmpty(request, "onSignOut message not pubs.SignOut")
	utils.AssertEmpty(sender, "onSignOut sender is null")

	accounts.SignOut(request.Token)

	var (
		uid     string
		err     error
		results map[string]string
	)

	uid, err = rds.GetPlayerUID(ctx, sender)
	if err != nil {
		return nil, err
	}

	if results, err = rds.UnRegisterPlayer(ctx, sender, uid); err != nil {
		actor.submitRequestCloseClient(ctx, sender)
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

func (actor *LoginActor) submitRequestCloseClient(ctx  context.Context, clientId *network.ClientID) {
	actor.submitRequest(ctx, &prvs.RequestGatewayCloseClient{Target: clientId})
}

func (actor *LoginActor) submitRequest(ctx  context.Context, request proto.Message) (proto.Message, error) {
	r := actor.ancestor.FindRouter(request)
	if r == nil {
		vlog.Errorf("%s unfound router", proto.MessageName(request))
		return nil, errors.New("unfound router")
	}

	result, err := r.Proxy.RequestMessage(request, defaultRequestTimeout)
	if err != nil {
		vlog.Errorf("%s fail error %s", proto.MessageName(request), err.Error())
	}

	return result, err
}

func (actor *LoginActor) Closed(ctx  context.Context) {

}
