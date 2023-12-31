package gateway

import (
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
func (actor *GatewayServantActor) onRequestGatewayPush(ctx *serve.ServantClientContext) (proto.Message, error) {
	backward := ctx.Message.(*prvs.RequestGatewayPush)
	utils.AssertEmpty(backward, "onBackwardBundle no prvs.RequestGatewayPush")

	if backward.Target == nil {
		return nil, errors.New("RequestGatewayPush target is null")
	}

	c := actor.gateway.GetClient(backward.Target)
	if c == nil {
		return nil, fmt.Errorf("RequestGatewayPush unfound client %s", backward.Target.ToString())
	}

	defer actor.gateway.ReleaseClient(c)

	body := backward.Body.ProtoReflect().New().Interface()

	if err := backward.Body.UnmarshalTo(body); err != nil {
		panic(err)
	}

	return nil, c.Post(body)
}

// onAlterRule 客户端角色修改
func (actor *GatewayServantActor) onRequestGatewayAlterRule(ctx *serve.ServantClientContext) (proto.Message, error) {
	alter := ctx.Message.(*prvs.RequestGatewayAlterRule)
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
func (actor *GatewayServantActor) onRequestGatewayCloseClient(ctx *serve.ServantClientContext) (proto.Message, error) {
	closing := ctx.Message.(*prvs.RequestGatewayCloseClient)
	utils.AssertEmpty(closing, "onRequestGatewayCloseClient no prvs.RequestGatewayCloseClient")
	c := actor.gateway.GetClient(closing.Target)
	if c == nil {
		return nil, fmt.Errorf("RequestGatewayCloseClient unfound client %s", closing.Target.ToString())
	}
	defer actor.gateway.ReleaseClient(c)
	c.ClientID().UserClose()

	return nil, nil
}

func (actor *GatewayServantActor) Closed(ctx *serve.ServantClientContext) {

}
