package apps

import (
	"context"
	"errors"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	"github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"

	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	// mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	// "github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/vlog"
	// "google.golang.org/protobuf/reflect/protoreflect"
	// "google.golang.org/protobuf/types/known/anypb"
)

const (
	defaultRequestTimeout = 1000
)

type LoginActor struct {
	ancestor *serve.Servant
}

func (actor *LoginActor) OnSignIn(ctx context.Context, req *pubs.SignIn) (_r *pubs.SignInResp, _err error) {
	// request := serve.GetServantClientInfo(ctx).Message().(*mpubs.SignIn)
	sender := serve.GetServantClientInfo(ctx).Sender()
	// utils.AssertEmpty(request, "onSignIn message not protocols.SignIn")
	utils.AssertEmpty(sender, "onSignIn sender is null")

	player, err := accounts.SignIn(ctx, req.Token)
	if err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.ContextDebugf(ctx, "onSignin error %s", err.Error())
		return nil, err
	}

	if err := rds.RegisterPlayer(ctx, sender,
		player.UID, player.DisplayName, player); err != nil {
		// actor.submitRequestCloseClient(ctx, sender)
		return nil, err
	}
	actor.submitRequestGatewayAlterRule(ctx, sender,3)
	resp := &pubs.SignInResp{}
	resp.UID = player.UID
	resp.DisplayName = player.DisplayName
	resp.Externs = make(map[string]string)
	for k, v := range player.Externs {
		resp.Externs[k] = v
	}
	return resp, nil
}

func (actor *LoginActor) OnSignOut(ctx context.Context, req *pubs.SignOut) (_r *pubs.SignOutResp, _err error) {

	// request := serve.GetServantClientInfo(ctx).Message().(*mpubs.SignOut)
	sender := serve.GetServantClientInfo(ctx).Sender()
	// utils.AssertEmpty(request, "onSignOut message not pubs.SignOut")
	utils.AssertEmpty(sender, "onSignOut sender is null")

	// accounts.SignOut(request.Token)

	var (
		// uid     string
		// err     error
		// results map[string]string
	)

	// uid, err = rds.GetPlayerUID(ctx, sender)
	// if err != nil {
	// 	return nil, err
	// }

	// results, err = rds.FindPlayerData(ctx, sender)
	// if err != nil {
	// 	return nil, err
	// }
	// if results != nil {
	// 	// TODO:退出Battle
	// 	if battleSpaceId, ok := results[rdsconst.PlayerMapClientBattleSpaceId]; ok && battleSpaceId != "" {
	// 		actor.submitRequest(ctx, &mprvs.RequestExitBattleSpace{BattleSpaceID: battleSpaceId, UID: uid})
	// 		// TODO: 是否需要判断执行失败
	// 	}
	// }

	// if err = rds.UnRegisterPlayer(ctx, sender); err != nil {
	// 	// actor.submitRequestCloseClient(ctx, sender)
	// 	return nil, err
	// }
	return nil,nil
	// return &mpubs.SignOutResp{
	// 	Result: 0,
	// }, nil
}

func (actor *LoginActor)  OnClientClosed(ctx context.Context, req *prvs.ClientClosed) (_err error){
	request := serve.GetServantClientInfo(ctx).Message().(*prvs.ClientClosed)
	_, err := rds.GetPlayerUID(ctx, request.ClientID)
	if err != nil {
		return err
	}

	results, err := rds.FindPlayerData(ctx, request.ClientID)
	if err != nil {
		return  err
	}
	if results != nil {
		// TODO:退出Battle
		if battleSpaceId, ok := results[rdsconst.PlayerMapClientBattleSpaceId]; ok && battleSpaceId != "" {
			// actor.submitForwardBundleRequest(ctx,request.ClientID, &mprvs.RequestExitBattleSpace{BattleSpaceID: battleSpaceId, UID: uid})
			// TODO: 是否需要判断执行失败
		}
	}
	if err := rds.UnRegisterPlayer(ctx, request.ClientID); err != nil {
		return  err
	}
	return nil
}

func (actor *LoginActor) submitRequestCloseClient(ctx context.Context, clientId *network.ClientID) {
	actor.submitRequest(ctx, &prvs.RequestGatewayCloseClient{Target: clientId})
}

func (actor *LoginActor) submitRequestGatewayAlterRule(ctx context.Context, clientId *network.ClientID,rule int32){
	actor.submitRequest(ctx, &prvs.RequestGatewayAlterRule{Target: clientId,Rule: rule})
}

func (actor *LoginActor) submitForwardBundleRequest(ctx context.Context,clientId *network.ClientID, request thrift.TStruct)([]byte, error) {
	r := actor.ancestor.FindRouter(protocol.MessageName(request))
	if r == nil {
		vlog.Errorf("%s unfound router", protocol.MessageName(request))
		return nil, errors.New("unfound router")
	}
	oprot := protocol.NewBinaryProtocol()
	bodyAny ,err := messages.MarshalTStruct(ctx, oprot, request,serve.GetServantClientInfo(ctx).SeqId())
	if err != nil {
		vlog.Warnf("%s message encoding failed error %s",
		protocol.MessageName(request), err.Error())
		return nil, err
	}

	forwardBundle := &prvs.ForwardBundle{
		Sender: clientId,
		Body:   bodyAny,
	}

	// 采用平均时间
	result, err := r.Proxy.RequestMessage(forwardBundle, defaultRequestTimeout)
	if err != nil {
		vlog.Errorf("%s fail error %s", protocol.MessageName(request), err.Error())
	}

	return result, err
}

func (actor *LoginActor) submitRequest(ctx context.Context, request thrift.TStruct) ([]byte, error) {
	r := actor.ancestor.FindRouter(protocol.MessageName(request))
	if r == nil {
		vlog.Errorf("%s unfound router",protocol.MessageName(request))
		return nil, errors.New("unfound router")
	}

	// 采用平均时间
	result, err := r.Proxy.RequestMessage(request, defaultRequestTimeout)
	if err != nil {
		vlog.Errorf("%s fail error %s", protocol.MessageName(request), err.Error())
	}

	return result, err
}

func (actor *LoginActor) Closed(ctx context.Context) {

}
