package gateway

import (
	"context"
	"errors"
	"fmt"

	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/utils"
	"google.golang.org/protobuf/proto"
)

type GatewayServantActor struct {
	gateway *Gateway
}

// onBackwardBundle 回退转发包
func (actor *GatewayServantActor) onRequestGatewayPush(ctx context.Context) (proto.Message, error) {
	backward := serve.GetServantClientInfo(ctx).Message().(*prvs.RequestGatewayPush)
	utils.AssertEmpty(backward, "onBackwardBundle no prvs.RequestGatewayPush")

	if backward.Target == nil {
		return nil, errors.New("RequestGatewayPush target is null")
	}

	c := actor.gateway.GetClient(backward.Target)
	if c == nil {
		return nil, fmt.Errorf("RequestGatewayPush unfound client %s", backward.Target.ToString())
	}

	defer actor.gateway.ReleaseClient(c)
	body, err := backward.Body.UnmarshalNew()
	if err != nil {
		panic(err)
	}

	return nil, c.Post(body)
}

// onAlterRule 客户端角色修改
func (actor *GatewayServantActor) onRequestGatewayAlterRule(ctx context.Context) (proto.Message, error) {
	alter := serve.GetServantClientInfo(ctx).Message().(*prvs.RequestGatewayAlterRule)
	utils.AssertEmpty(alter, "onRequestGatewayAlterRule no prvs.RequestGatewayAlterRule")

	c := actor.gateway.GetClient(alter.Target)
	if c == nil {
		return nil, fmt.Errorf("RequestGatewayAlterRule unfound client %s", alter.Target.ToString())
	}
	defer actor.gateway.ReleaseClient(c)
	c.alterRule(alter.Rule)

	return nil, nil
}

// onClientCloseRequest 关闭请求
func (actor *GatewayServantActor) onRequestGatewayCloseClient(ctx context.Context) (proto.Message, error) {
	closing := serve.GetServantClientInfo(ctx).Message().(*prvs.RequestGatewayCloseClient)
	utils.AssertEmpty(closing, "onRequestGatewayCloseClient no prvs.RequestGatewayCloseClient")
	c := actor.gateway.GetClient(closing.Target)
	if c == nil {
		return nil, fmt.Errorf("RequestGatewayCloseClient unfound client %s", closing.Target.ToString())
	}
	// vlog.Infof("RequestGatewayCloseClient : ", closing.Target.ToString())
	defer actor.gateway.ReleaseClient(c)
	go c.ClientID().UserClose()

	return closing, nil
}

func (actor *GatewayServantActor) Closed(ctx context.Context) {

}
