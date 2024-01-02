package apps

import (
	"errors"
	"strconv"

	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/dba/rds"
	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	"github.com/yamakiller/velcro-go/network"
	"google.golang.org/protobuf/proto"
)

const (
	defaultRequestTimeout = 1000
)

type BattleActor struct {
	ancestor *serve.Servant
}

func (actor *BattleActor) onCreateBattleSpace(ctx *serve.ServantClientContext) (proto.Message, error) {
	request := ctx.Message.(*mpubs.CreateBattleSpace)
	var (
		uid     string
		err     error
		spaceid string
	)
	uid, err = rds.GetPlayerUID(ctx.Background, ctx.Sender)
	if err != nil {
		actor.submitRequestCloseClient(ctx, ctx.Sender)
		ctx.Context.Debug("onCreateBattleSpace error %s", err.Error())
		return nil, err
	}
	//TODO: 创建战场
	spaceid, err = rds.CreateBattleSpace(ctx.Background, uid, ctx.Sender, request.MapURI)
	if err != nil {
		actor.submitRequestCloseClient(ctx, ctx.Sender)
		ctx.Context.Debug("onCreateBattleSpace error %s", err.Error())
		return nil, err
	}
	res := &mpubs.CreateBattleSpaceResp{}
	res.MapURI = request.MapURI
	res.SpaceId = spaceid
	return res, nil
}

func (actor *BattleActor) onGetBattleSpaceList(ctx *serve.ServantClientContext) (proto.Message, error) {
	request := ctx.Message.(*mpubs.GetBattleSpaceList)
	var (
		spaceids []string
		err      error
		total    int64
	)
	total, err = rds.GetBattleSpacesCount(ctx.Background)
	if err != nil {
		actor.submitRequestCloseClient(ctx, ctx.Sender)
		ctx.Context.Debug("onGetBattleSpaceList error %s", err.Error())
		return nil, err
	}
	spaceids, err = rds.GetBattleSpaceList(ctx.Background, int64(request.Start), int64(request.Size))
	if err != nil {
		actor.submitRequestCloseClient(ctx, ctx.Sender)
		ctx.Context.Debug("onGetBattleSpaceList error %s", err.Error())
		return nil, err
	}
	res := &mpubs.GetBattleSpaceListResp{}
	res.Count = int32(total)
	res.Start = request.Start
	for _, spaceid := range spaceids {
		result, err := rds.FindBattleSpaceData(ctx.Background, spaceid)
		if err != nil {
			continue
		}
		spacse := &mpubs.BattleSpaceDataSimple{}
		spacse.SpaceId = spaceid
		spacse.MapURI = result[rdsconst.BattleSpaceMapURi]
		spacse.MasterUid = result[rdsconst.BattleSpaceMasterUid]
		spacse.MasterIcon = result[rdsconst.BattleSpaceMasterIcon]
		spacse.MasterDisplay = result[rdsconst.BattleSpaceMasterDisplay]
		if result[rdsconst.BattleSpacePos1Uid] != "" {
			player := &mpubs.BattleSpacePlayerSimple{

				Display: result[rdsconst.BattleSpacePos1Display],
				Pos:     0,
			}
			spacse.Players = append(spacse.Players, player)
		}
	
		if result[rdsconst.BattleSpacePos2Uid] != "" {
			player := &mpubs.BattleSpacePlayerSimple{
				Display: result[rdsconst.BattleSpacePos2Display],
				Pos:     1,
			}
			spacse.Players = append(spacse.Players, player)
		}
	
		if result[rdsconst.BattleSpacePos3Uid] != "" {
			player := &mpubs.BattleSpacePlayerSimple{
				Display: result[rdsconst.BattleSpacePos3Display],
				Pos:     2,
			}
			spacse.Players = append(spacse.Players, player)
		}
	
		if result[rdsconst.BattleSpacePos4Uid] != "" {
			player := &mpubs.BattleSpacePlayerSimple{

				Display: result[rdsconst.BattleSpacePos4Display],
				Pos:     3,
			}
			spacse.Players = append(spacse.Players, player)
		}
		res.Spaces = append(res.Spaces, spacse)
	}
	return nil, nil
}

func (actor *BattleActor) onEnterBattleSpace(ctx *serve.ServantClientContext) (proto.Message, error) {
	request := ctx.Message.(*mpubs.EnterBattleSpace)
	
	if err := rds.EnterBattleSpace(ctx.Background, request.SpaceId, ctx.Sender); err != nil {
		actor.submitRequestCloseClient(ctx, ctx.Sender)
		ctx.Context.Debug("onEnterBattleSpace error %s", err.Error())
		return nil, err
	}
	result, err := rds.FindBattleSpaceData(ctx.Background, request.SpaceId)
	if err != nil {
		actor.submitRequestCloseClient(ctx, ctx.Sender)
		ctx.Context.Debug("onEnterBattleSpace error %s", err.Error())
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

	if result[rdsconst.BattleSpacePos1Uid] != "" {
		player := &mpubs.BattleSpacePlayer{
			Uid:     result[rdsconst.BattleSpacePos1Uid],
			Icon:    result[rdsconst.BattleSpacePos1Icon],
			Display: result[rdsconst.BattleSpacePos1Display],
			Pos:     0,
		}
		res.Space.Players = append(res.Space.Players, player)
	}

	if result[rdsconst.BattleSpacePos2Uid] != "" {
		player := &mpubs.BattleSpacePlayer{
			Uid:     result[rdsconst.BattleSpacePos2Uid],
			Icon:    result[rdsconst.BattleSpacePos2Icon],
			Display: result[rdsconst.BattleSpacePos2Display],
			Pos:     1,
		}
		res.Space.Players = append(res.Space.Players, player)
	}

	if result[rdsconst.BattleSpacePos3Uid] != "" {
		player := &mpubs.BattleSpacePlayer{
			Uid:     result[rdsconst.BattleSpacePos3Uid],
			Icon:    result[rdsconst.BattleSpacePos3Icon],
			Display: result[rdsconst.BattleSpacePos3Display],
			Pos:     2,
		}
		res.Space.Players = append(res.Space.Players, player)
	}

	if result[rdsconst.BattleSpacePos4Uid] != "" {
		player := &mpubs.BattleSpacePlayer{
			Uid:     result[rdsconst.BattleSpacePos4Uid],
			Icon:    result[rdsconst.BattleSpacePos4Icon],
			Display: result[rdsconst.BattleSpacePos4Display],
			Pos:     3,
		}
		res.Space.Players = append(res.Space.Players, player)
	}

	return res, nil
}

func (actor *BattleActor) onDisplayBattleSpace(ctx *serve.ServantClientContext)(proto.Message, error){
	request := ctx.Message.(*mpubs.DisplayBattleSpace)
	if err := rds.DisplayBattleSpace(ctx.Background, request.SpaceId, request.Uid, request.Display,ctx.Sender); err != nil {
		actor.submitRequestCloseClient(ctx, ctx.Sender)
		ctx.Context.Debug("onRequestExitBattleSpace error %s", err.Error())
		return nil, err
	}
	res := &mpubs.DisplayBattleSpaceResp{}
	res.SpaceId = request.SpaceId
	res.Uid = request.Uid
	res.Display = request.Display
	return res,nil
}
func (actor *BattleActor) onReportNat(ctx *serve.ServantClientContext) (proto.Message, error) {
	return nil, nil
}

func (actor *BattleActor) onRequestExitBattleSpace(ctx *serve.ServantClientContext) (proto.Message, error) {
	request := ctx.Message.(*mprvs.RequestExitBattleSpace)
	if err := rds.LeaveBattleSpace(ctx.Background,request.BattleSpaceID,request.UID,ctx.Sender); err != nil{
		actor.submitRequestCloseClient(ctx, ctx.Sender)
		ctx.Context.Debug("onRequestExitBattleSpace error %s", err.Error())
		return nil, err
	}
	res := &mprvs.RequestExitBattleSpace{}

	return res, nil
}

func (actor *BattleActor) submitRequestCloseClient(ctx *serve.ServantClientContext, clientId *network.ClientID) {
	actor.submitRequest(ctx, &prvs.RequestGatewayCloseClient{Target: clientId})
}

func (actor *BattleActor) submitRequest(ctx *serve.ServantClientContext, request proto.Message) (proto.Message, error) {
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

func (actor *BattleActor) Closed(ctx *serve.ServantClientContext) {

}
