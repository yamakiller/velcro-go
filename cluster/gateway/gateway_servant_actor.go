package gateway

import (
	"context"
	"errors"
	"fmt"

	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/utils"
)

type GatewayServantActor struct {
	gateway *Gateway
}
// OnRequestGatewayPush 回退转发包
func (actor *GatewayServantActor) OnRequestGatewayPush(ctx context.Context,req *prvs.RequestGatewayPush) (res *prvs.RequestGatewayPush,_err error){
	backward := serve.GetServantClientInfo(ctx).Message().(*prvs.RequestGatewayPush)
	utils.AssertEmpty(backward, "onBackwardBundle no prvs.RequestGatewayPush")

	if backward.Target == nil {
		return nil, errors.New("RequestGatewayPush target is null")
	}

	c := actor.gateway.GetClient(backward.Target)
	if c == nil {
		return nil,fmt.Errorf("RequestGatewayPush unfound client %s", backward.Target.ToString())
	}

	return req,c.Post(backward.Body)
}
// OnRequestGatewayAlterRule 客户端角色修改
func (actor *GatewayServantActor) OnRequestGatewayAlterRule(ctx context.Context, req *prvs.RequestGatewayAlterRule) (res *prvs.RequestGatewayAlterRule,_err error){
	c := actor.gateway.GetClient(req.Target)
	if c == nil {
		return nil, fmt.Errorf("RequestGatewayAlterRule unfound client %s", req.Target.String())
	}
	defer actor.gateway.ReleaseClient(c)
	c.alterRule(req.Rule)
	return req, nil
}

// onClientCloseRequest 关闭请求
func (actor *GatewayServantActor) OnRequestGatewayCloseClient(ctx context.Context,req *prvs.RequestGatewayCloseClient) (res *prvs.RequestGatewayCloseClient,_err error){

	c := actor.gateway.GetClient(req.Target)
	if c == nil {
		return nil, fmt.Errorf("RequestGatewayCloseClient unfound client %s", req.Target.ToString())
	}
	// vlog.Infof("RequestGatewayCloseClient : ", req.Target.ToString())
	defer actor.gateway.ReleaseClient(c)
	go c.ClientID().UserClose()

	return req, nil
}

func (actor *GatewayServantActor) Closed(ctx context.Context) {

}
