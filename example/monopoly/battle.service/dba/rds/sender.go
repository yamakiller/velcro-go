package rds

import (
	"context"
	"errors"

	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/serve"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	ancestor *serve.Servant
)

func WithServant(ser *serve.Servant) {
	ancestor = ser
}

func sendDisRoomWarningNotify(spaceid string, tts int64) {
	res := &mpubs.DisRoomWarningNotify{
		SpaceID: spaceid,
		Tts:     tts,
	}
	players := GetBattleSpacePlayers(context.Background(), spaceid)
	for _, v := range players {
		submitRequestGatewayPush(context.Background(), v, res)
	}
}

func sendDissBattleSpaceNotify(spaceid string, clientId *network.ClientID) {
	res := &mpubs.DissBattleSpaceNotify{
		SpaceId: spaceid,
	}
	submitRequestGatewayPush(context.Background(), clientId, res)
}

func submitRequestGatewayPush(ctx context.Context, clientId *network.ClientID, request proto.Message) error {
	r := ancestor.FindAddrRouter(clientId.Address)
	if r == nil {
		if ctx != nil {
			vlog.Errorf("%s unfound router", proto.MessageName(request))
		}
		return errors.New("unfound router")
	}
	dataAny, _ := anypb.New(request)
	bounld := &prvs.RequestGatewayPush{
		Target: clientId,
		Body:   dataAny,
	}
	_, err := r.Proxy.RequestMessage(bounld, 1000)
	if err != nil {
		vlog.Errorf("%s fail error %s", proto.MessageName(request), err.Error())
	}
	return nil
}
