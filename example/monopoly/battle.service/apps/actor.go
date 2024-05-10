package apps

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/dba/rds"
	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	defaultRequestTimeout = 1000
)

type BattleActor struct {
	ancestor *serve.Servant
}

func (actor *BattleActor) onCreateBattleSpace(ctx context.Context) (proto.Message, error) {

	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.CreateBattleSpace)
	sender := serve.GetServantClientInfo(ctx).Sender()
	var (
		uid     string
		err     error
		spaceid string
	)
	uid, err = rds.GetPlayerUID(ctx, sender)
	if err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onCreateBattleSpace error %s", err.Error())
		return nil, err
	}
	//TODO: 创建战场
	spaceid, err = rds.CreateBattleSpace(ctx, uid, sender, request.MapURI, int32(request.MaxCount))
	if err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onCreateBattleSpace error %s", err.Error())
		return nil, err
	}
	res := &mpubs.CreateBattleSpaceResp{}
	res.MapURI = request.MapURI
	res.SpaceId = spaceid
	return res, nil
}

func (actor *BattleActor) onGetBattleSpaceList(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.GetBattleSpaceList)
	sender := serve.GetServantClientInfo(ctx).Sender()

	var (
		spaceids []string
		err      error
		total    int64
	)
	total, err = rds.GetBattleSpacesCount(ctx)
	if err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onGetBattleSpaceList error %s", err.Error())
		return nil, err
	}
	spaceids, err = rds.GetBattleSpaceList(ctx, int64(request.Start), int64(request.Size))
	if err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onGetBattleSpaceList error %s", err.Error())
		return nil, err
	}
	res := &mpubs.GetBattleSpaceListResp{}
	res.Count = int32(total)
	res.Start = request.Start
	for _, spaceid := range spaceids {
		result, err := rds.FindBattleSpaceData(ctx, spaceid)
		if err != nil {
			continue
		}
		spacse := &mpubs.BattleSpaceDataSimple{}
		spacse.SpaceId = spaceid
		spacse.MapURI = result[rdsconst.BattleSpaceMapURi]
		spacse.MasterUid = result[rdsconst.BattleSpaceMasterUid]
		spacse.MasterIcon = result[rdsconst.BattleSpaceMasterIcon]
		spacse.MasterDisplay = result[rdsconst.BattleSpaceMasterDisplay]
		max_count ,_ := strconv.Atoi(result[rdsconst.BattleSpacePlayerCount]) 
		spacse.MaxCount =int32(max_count) 
		battleSpacePos := result[rdsconst.BattleSpacePlayerPos]
		for _, id := range strings.Split(battleSpacePos, "&") {
			player_data := rdsconst.SplitData(result[rdsconst.GetBattleSpacePlayerDataKey(id)])
			pos, _ := (strconv.Atoi(player_data[4]))
			player := &mpubs.BattleSpacePlayerSimple{
				Display: player_data[2],
				Pos:     int32(pos),
			}
			spacse.Players = append(spacse.Players, player)
		}
		res.Spaces = append(res.Spaces, spacse)
	}
	return res, nil
}

func (actor *BattleActor) onEnterBattleSpace(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.EnterBattleSpace)
	sender := serve.GetServantClientInfo(ctx).Sender()

	if err := rds.EnterBattleSpace(ctx, request.SpaceId, sender); err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onEnterBattleSpace error %s", err.Error())
		return nil, err
	}
	result, err := rds.FindBattleSpaceData(ctx, request.SpaceId)
	if err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onEnterBattleSpace error %s", err.Error())
		return nil, err
	}
	res := &mpubs.EnterBattleSpaceResp{}
	res.Space = &mpubs.BattleSpaceData{}
	res.Space.SpaceId = result[rdsconst.BattleSpaceId]
	res.Space.MapURI = result[rdsconst.BattleSpaceMapURi]
	res.Space.MasterUid = result[rdsconst.BattleSpaceMasterUid]
	time, _ := strconv.Atoi(result[rdsconst.BattleSpaceTime])
	res.Space.Starttime = uint64(time)
	res.Space.State = result[rdsconst.BattleSpaceState]

	battleSpacePos := result[rdsconst.BattleSpacePlayerPos]
	for _, id := range strings.Split(battleSpacePos, "&") {
		if id == "" {
			continue
		}
		player_data := rdsconst.SplitData(result[rdsconst.GetBattleSpacePlayerDataKey(id)])
		pos, _ := (strconv.Atoi(player_data[4]))
		player := &mpubs.BattleSpacePlayer{
			Uid:     player_data[0],
			Icon:    player_data[1],
			Display: player_data[2],
			Pos:     int32(pos),
		}
		res.Space.Players = append(res.Space.Players, player)
	}
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceId)
	for _, v := range players {
		if !sender.Equal(v) {
			actor.submitRequestGatewayPush(ctx, v, res)
		}
	}
	return res, nil
}

func (actor *BattleActor) onReadyBattleSpace(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.ReadyBattleSpace)
	sender := serve.GetServantClientInfo(ctx).Sender()

	if err := rds.ReadyBattleSpace(ctx, request.SpaceId, request.Ready, sender); err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onRequestExitBattleSpace error %s", err.Error())
		return nil, err
	}
	res := &mpubs.ReadyBattleSpaceResp{}
	res.SpaceId = request.SpaceId
	res.Uid = request.Uid
	res.Ready = request.Ready
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceId)
	for _, v := range players {
		if !sender.Equal(v) {
			actor.submitRequestGatewayPush(ctx, v, res)
		}
	}
	return res, nil
}

func (actor *BattleActor) onRequestStartBattleSpace(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.RequsetStartBattleSpace)
	sender := serve.GetServantClientInfo(ctx).Sender()
	if err := rds.StartBattleSpace(ctx, request.SpaceId, sender); err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onRequestStartBattleSpace error %s", err.Error())
		return nil, err
	}
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceId)
	res := &mpubs.RequsetStartBattleSpaceResp{}
	res.SpaceId = request.SpaceId

	for _, v := range players {
		if !sender.Equal(v) {
			actor.submitRequestGatewayPush(ctx, v, res)
		}
	}
	return res, nil
}

func (actor *BattleActor) onReportNat(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mprvs.ReportNat)
	sender := serve.GetServantClientInfo(ctx).Sender()
	if err := rds.BattleSpaceReportNat(ctx, request.BattleSpaceID, request.NatAddr, sender); err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onReportNat error %s", err.Error())
		return nil, err
	}
	players := rds.GetBattleSpacePlayers(ctx, request.BattleSpaceID)
	for _, v := range players {
		actor.submitRequestGatewayPush(ctx, v, request)
	}
	return request, nil
}

func (actor *BattleActor) onRequestExitBattleSpace(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mprvs.RequestExitBattleSpace)
	sender := serve.GetServantClientInfo(ctx).Sender()
	
	// request := ctx.Message.(*mprvs.RequestExitBattleSpace)
	if ok, err := rds.IsMaster(ctx, sender); err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onRequestExitBattleSpace error %s", err.Error())
	} else if ok {
		// 房主离开，解散房间
		players := rds.GetBattleSpacePlayers(ctx, request.BattleSpaceID)
		if err := rds.DeleteBattleSpace(ctx, sender); err != nil {
			return nil, nil
		}

		res := &mpubs.DissBattleSpaceNotify{
			SpaceId: request.BattleSpaceID,
		}
		for _, v := range players {
			if !sender.Equal(v) {
				actor.submitRequestGatewayPush(ctx, v, res)
			}
		}
	} else {
		// 非房主离开，退出房间
		rds.LeaveBattleSpace(ctx, sender)
		res := &mprvs.RequestExitBattleSpace{
			BattleSpaceID: request.BattleSpaceID,
			UID:           request.UID,
		}
		players := rds.GetBattleSpacePlayers(ctx, request.BattleSpaceID)
		for _, v := range players {
			if !sender.Equal(v) {
				actor.submitRequestGatewayPush(ctx, v, res)
			}
		}
	}

	return &mprvs.RequestExitBattleSpace{
		BattleSpaceID: request.BattleSpaceID,
		UID:           request.UID,
	}, nil
}

func (actor *BattleActor) submitRequestGatewayPush(ctx context.Context, clientId *network.ClientID, request proto.Message) error {
	r := actor.ancestor.FindAddrRouter(clientId.Address)
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
	_, err := r.Proxy.RequestMessage(bounld, defaultRequestTimeout)
	if err != nil {
		vlog.Errorf("%s fail error %s", proto.MessageName(request), err.Error())
	}
	return nil
}

func (actor *BattleActor) submitRequestCloseClient(ctx context.Context, clientId *network.ClientID) {

	actor.submitRequest(ctx, &prvs.RequestGatewayCloseClient{Target: clientId})
}

func (actor *BattleActor) submitRequest(ctx context.Context, request proto.Message) (proto.Message, error) {
	r := actor.ancestor.FindRouter(request)
	if r == nil {
		if ctx != nil {
			vlog.Errorf("%s unfound router", proto.MessageName(request))
		}

		return nil, errors.New("unfound router")
	}

	result, err := r.Proxy.RequestMessage(request, defaultRequestTimeout)
	if err != nil {
		if ctx != nil {
			vlog.Errorf("%s fail error %s", proto.MessageName(request), err.Error())
		}
	}

	return result, err
}

func (actor *BattleActor) Closed(ctx context.Context) {
}
