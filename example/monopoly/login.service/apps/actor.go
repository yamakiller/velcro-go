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
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsstruct"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	defaultRequestTimeout = 1000
)

type LoginActor struct {
	ancestor *serve.Servant
}

func (actor *LoginActor) onSignIn(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.SignIn)
	sender := serve.GetServantClientInfo(ctx).Sender()
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
	actor.submitRequestGatewayAlterRule(ctx, sender, 3)
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
	sender := serve.GetServantClientInfo(ctx).Sender()
	utils.AssertEmpty(request, "onSignOut message not pubs.SignOut")
	utils.AssertEmpty(sender, "onSignOut sender is null")

	accounts.SignOut(request.Token)

	var (
		uid     string
		err     error
		results *rdsstruct.RdsPlayerData
	)

	uid, err = rds.GetPlayerUID(ctx, sender)
	if err != nil {
		return nil, err
	}

	results, err = rds.FindPlayerData(ctx, sender)
	if err != nil {
		return nil, err
	}
	vlog.Debugf("onSignOut %s", uid)
	if results != nil {
		// TODO:退出Battle
		BattleSpaceId, err := rds.GetPlayerBattleSpaceID(ctx, uid)
		if err == nil && BattleSpaceId != "" {
			actor.submitRequest(ctx, &mprvs.RequestExitBattleSpace{BattleSpaceID: BattleSpaceId, UID: uid})
			// TODO: 是否需要判断执行失败
		}
	}

	if err = rds.UnRegisterPlayer(ctx, sender); err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		return nil, err
	}

	return &mpubs.SignOutResp{
		Result: 0,
	}, nil
}

func (actor *LoginActor) onClientClosed(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*prvs.ClientClosed)
	uid, err := rds.GetPlayerUID(ctx, request.ClientID)
	if err != nil {
		return nil, err
	}
	results, err := rds.FindPlayerData(ctx, request.ClientID)
	if err != nil {
		return nil, err
	}
	if results != nil {
		// TODO:退出Battle
		BattleSpaceId, err := rds.GetPlayerBattleSpaceID(ctx, uid)
		if err == nil && BattleSpaceId != "" {
			actor.submitForwardBundleRequest(ctx, request.ClientID, &mprvs.RequestExitBattleSpace{BattleSpaceID: BattleSpaceId, UID: uid})
			// TODO: 是否需要判断执行失败
		}
	}
	if err := rds.UnRegisterPlayer(ctx, request.ClientID); err != nil {
		return nil, err
	}
	return nil, nil
}

func (actor *LoginActor) submitRequestCloseClient(ctx context.Context, clientId *network.ClientID) {
	actor.submitRequest(ctx, &prvs.RequestGatewayCloseClient{Target: clientId})
}

func (actor *LoginActor) submitRequestGatewayAlterRule(ctx context.Context, clientId *network.ClientID, rule int32) {
	actor.submitRequest(ctx, &prvs.RequestGatewayAlterRule{Target: clientId, Rule: rule})
}

func (actor *LoginActor) submitForwardBundleRequest(_ context.Context, clientId *network.ClientID, request proto.Message) (proto.Message, error) {
	r := actor.ancestor.FindRouter(request)
	if r == nil {
		vlog.Errorf("%s unfound router", proto.MessageName(request))
		return nil, errors.New("unfound router")
	}
	bodyAny, err := anypb.New(request)
	if err != nil {
		vlog.Warnf("%s message encoding failed error %s",
			string(protoreflect.FullName(proto.MessageName(request))), err.Error())
		return nil, err
	}

	forwardBundle := &prvs.ForwardBundle{
		Sender: clientId,
		Body:   bodyAny,
	}

	// 采用平均时间
	result, err := r.Proxy.RequestMessage(forwardBundle, defaultRequestTimeout)
	if err != nil {
		vlog.Errorf("%s fail error %s", proto.MessageName(request), err.Error())
	}

	return result, err
}

func (actor *LoginActor) submitRequest(_ context.Context, request proto.Message) (proto.Message, error) {
	r := actor.ancestor.FindRouter(request)
	if r == nil {
		vlog.Errorf("%s unfound router", proto.MessageName(request))
		return nil, errors.New("unfound router")
	}

	// 采用平均时间
	result, err := r.Proxy.RequestMessage(request, defaultRequestTimeout)
	if err != nil {
		vlog.Errorf("%s fail error %s", proto.MessageName(request), err.Error())
	}

	return result, err
}

func (actor *LoginActor) Closed(ctx context.Context) {

}
