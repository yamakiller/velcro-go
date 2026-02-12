package apps

import (
	"context"
	"errors"
	"time"

	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/dba/rds"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/errs"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/jwt"
	mprvs "github.com/yamakiller/velcro-go/example/monopoly/protocols/prvs"
	mpubs "github.com/yamakiller/velcro-go/example/monopoly/protocols/pubs"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsstruct"
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
		player  *rdsstruct.RdsPlayerData
		err     error
		spaceid string
	)
	player, err = rds.GetPlayerData(ctx, sender)
	if err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onCreateBattleSpace error %s", err.Error())
		return nil, err
	}
	//TODO: 创建战场
	spaceid, err = rds.CreateBattleSpace(ctx, player, sender, request.MapURI, int32(request.MaxCount), request.RoomName, request.Password, request.Extend, request.Display)
	if err != nil {
		// actor.submitRequestCloseClient(ctx, sender)
		// vlog.Debugf("onCreateBattleSpace error %s", err.Error())
		return nil, err
	}
	res := &mpubs.CreateBattleSpaceResp{}
	res.MapURI = request.MapURI
	res.SpaceId = spaceid
	return res, nil
}

func (actor *BattleActor) onGetBattleSpaceList(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.GetBattleSpaceList)
	// sender := serve.GetServantClientInfo(ctx).Sender()

	var (
		spaceids []string
		err      error
		total    int64
	)
	total, err = rds.GetBattleSpacesCount(ctx)
	if err != nil {
		return nil, err
	}
	spaceids, err = rds.GetBattleSpaceList(ctx, int64(request.Start), int64(request.Size))
	if err != nil {
		return nil, err
	}
	res := &mpubs.GetBattleSpaceListResp{}
	res.Count = int32(total)
	res.Start = request.Start
	for _, spaceid := range spaceids {
		result, err := rds.GetBattleSpaceInfo(ctx, spaceid)
		if err != nil || result == nil {
			rds.RemoveBattleSpaceList(spaceid)
			continue
		}

		spacse := &mpubs.BattleSpaceDataSimple{}
		spacse.SpaceId = spaceid
		spacse.RoomName = result.SpaceName
		spacse.IsHavePassword = (result.SpacePassword != "")
		spacse.MapURI = result.SpaceMapURI
		spacse.MasterUid = result.SpaceMasterUid
		spacse.MasterIcon = result.SpaceMasterIcon
		spacse.MasterDisplay = result.SpaceMasterDisplay
		spacse.MaxCount = int32(len(result.SpacePlayers))
		spacse.Players = make([]*mpubs.BattleSpacePlayerSimple, 0)
		for _, v := range result.SpacePlayers {
			if v.Uid != "" {
				spacse.Players = append(spacse.Players, &mpubs.BattleSpacePlayerSimple{
					Display: v.Display,
					Pos:     v.Pos,
				})
			}
		}

		res.Spaces = append(res.Spaces, spacse)
	}
	return res, nil
}

func (actor *BattleActor) onEnterBattleSpace(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.EnterBattleSpace)
	sender := serve.GetServantClientInfo(ctx).Sender()
	player_data, err := rds.GetPlayerData(ctx, sender)
	if err != nil {
		actor.submitRequestCloseClient(ctx, sender)
		vlog.Debugf("onEnterBattleSpace error %s", err.Error())
		return nil, err
	}

	if err := rds.EnterBattleSpace(ctx, request.SpaceId, request.Password, request.Display, sender); err != nil {
		// actor.submitRequestCloseClient(ctx, sender)
		// vlog.Debugf("onEnterBattleSpace error %s", err.Error())
		return nil, err
	}
	result, err := rds.GetBattleSpaceInfo(ctx, request.SpaceId)
	if err != nil {
		// actor.submitRequestCloseClient(ctx, sender)
		// vlog.Debugf("onEnterBattleSpace error %s", err.Error())
		return nil, err
	}
	res := &mpubs.EnterBattleSpaceResp{}
	res.Uid = player_data.UID
	res.Space = &mpubs.BattleSpaceData{}
	res.Space.SpaceId = result.SpaceId
	res.Space.MapURI = result.SpaceMapURI
	res.Space.MasterUid = result.SpaceMasterUid
	res.Space.Starttime = uint64(result.SpaceStarttime)
	res.Space.State = result.SpaceState
	res.Space.Extent = result.SpaceExtend
	res.Space.RoomName = result.SpaceName

	notify := &mpubs.EnterBattleSpaceNotify{
		SpaceId: request.SpaceId,
	}
	for _, v := range result.SpacePlayers {
		if v.Uid == "" {
			continue
		}
		player := &mpubs.BattleSpacePlayer{
			Uid:     v.Uid,
			Icon:    v.Icon,
			Display: v.Display,
			Pos:     v.Pos,
			Role:    v.Role,
			Camp:    v.Camp,
			Ready:   v.Ready,
			Extends: make(map[string]string),
			IsRobot: v.IsRobot,
		}

		for k, v := range v.Extends {
			player.Extends[k] = v
		}

		res.Space.Players = append(res.Space.Players, player)
		if player.Uid == player_data.UID {
			notify.Player = player
		}
	}
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceId)
	for _, v := range players {
		if !sender.Equal(v) {
			actor.submitRequestGatewayPush(ctx, v, notify)
		}
	}
	return res, nil
}

func (actor *BattleActor) onReadyBattleSpace(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.ReadyBattleSpace)
	sender := serve.GetServantClientInfo(ctx).Sender()

	if err := rds.ReadyBattleSpace(ctx, request.SpaceId, request.Ready, sender); err != nil {
		// actor.submitRequestCloseClient(ctx, sender)
		// vlog.Debugf("onRequestExitBattleSpace error %s", err.Error())
		return nil, err
	}
	notify := &mpubs.ReadyBattleSpaceNotify{}
	notify.SpaceId = request.SpaceId
	notify.Uid = request.Uid
	notify.Ready = request.Ready
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceId)
	for _, v := range players {
		if !sender.Equal(v) {
			actor.submitRequestGatewayPush(ctx, v, notify)
		}
	}
	res := &mpubs.ReadyBattleSpaceResp{}
	res.SpaceId = request.SpaceId
	res.Uid = request.Uid
	res.Ready = request.Ready
	return res, nil
}

func (actor *BattleActor) onRequestStartBattleSpace(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.RequsetStartBattleSpace)
	sender := serve.GetServantClientInfo(ctx).Sender()
	err := rds.IsBattleSpaceMaster(ctx, sender, request.SpaceId)
	if err != nil {
		return nil, err
	}

	// if err := rds.StartBattleSpace(ctx, request.SpaceId, sender); err != nil {
	// 	// actor.submitRequestCloseClient(ctx, sender)
	// 	// vlog.Debugf("onRequestStartBattleSpace error %s", err.Error())
	// 	return nil, err
	// }
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceId)
	res := &mpubs.RequsetStartBattleSpaceResp{
		Tokens: make(map[string]string),
	}

	spaceinfo, err := rds.GetBattleSpaceInfo(ctx, request.SpaceId)
	if err != nil {
		return nil, err
	}
	master := spaceinfo.SpaceMasterUid
	res.SpaceId = request.SpaceId
	auth := int32(0)
	for _, v := range players {
		player_data, err := rds.GetPlayerData(ctx, v)
		if err != nil {
			return nil, err
		}
		auth = 0
		if player_data.UID == master {
			auth = 1
		}
		expiration := time.Duration(envs.Instance().Get("configs").(*configs.Config).Server.JwtTimeout) * time.Second
		token, err := jwt.GenToken(envs.Instance().Get("configs").(*configs.Config).Server.JwtSecret, request.SpaceId, player_data.UID, auth, int32(len(spaceinfo.SpacePlayers)), expiration)
		if err != nil {
			return nil, err
		}
		res.Tokens[player_data.UID] = token
	}
	for _, v := range players {
		if !sender.Equal(v) {
			actor.submitRequestGatewayPush(ctx, v, res)
		}
	}

	rds.DeleteBattleSpace(ctx, sender)

	return res, nil
}

func (actor *BattleActor) onChanggePasswordRequest(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.ChanggePasswordRequest)
	sender := serve.GetServantClientInfo(ctx).Sender()

	if err := rds.IsMaster(ctx, sender); err != nil {
		return nil, err
	}

	if err := rds.ChangeBattleSpacePassword(ctx, request.SpaceID, request.NewPassword); err != nil {
		return nil, err
	}
	res := &mpubs.ChanggePasswordResponse{
		Res: 0,
	}
	return res, nil
}

func (actor *BattleActor) onKickUserRequest(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.KickUserRequest)
	sender := serve.GetServantClientInfo(ctx).Sender()
	if err := rds.IsMaster(ctx, sender); err != nil {
		return nil, err
	}
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceID)

	user := rds.GetBattleSpacePlayerClientID(ctx, request.Uid)
	if user == nil {
		return nil, errs.ErrorPlayerIsNotInBattleSpace
	}

	rds.LeaveBattleSpace(ctx, user)

	notify := &mpubs.UserExitSpaceNotify{
		SpaceID: request.SpaceID,
		Uid:     request.Uid,
	}
	for _, v := range players {
		actor.submitRequestGatewayPush(ctx, v, notify)
	}

	res := &mpubs.KickUserResponse{
		Res: 0,
	}
	return res, nil
}

func (actor *BattleActor) onModifyRoomParametersRequset(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.ModifyRoomParametersRequset)
	sender := serve.GetServantClientInfo(ctx).Sender()

	if err := rds.IsMaster(ctx, sender); err != nil {
		return nil, err
	}

	if err := rds.ChangeModifyRoomParameters(ctx, request.SpaceID, request.MapURI, request.MaxCount, request.RoomName, request.Extend); err != nil {
		vlog.Debugf("onModifyRoomParametersRequset ChangeModifyRoomParameters  %v", err.Error())
		return nil, err
	}

	players := rds.GetBattleSpacePlayers(ctx, request.SpaceID)
	notify := &mpubs.ModifyRoomParametersNotify{
		SpaceID:  request.SpaceID,
		MapURI:   request.MapURI,
		MaxCount: request.MaxCount,
		RoomName: request.RoomName,
		Extend:   request.Extend,
	}

	for _, v := range players {
		actor.submitRequestGatewayPush(ctx, v, notify)
	}

	res := &mpubs.ModifyRoomParametersResponse{
		Res: 0,
	}
	return res, nil
}

func (actor *BattleActor) onModifyUserRoleRequest(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.ModifyUserRoleRequest)
	sender := serve.GetServantClientInfo(ctx).Sender()

	player_data, err := rds.GetPlayerData(ctx, sender)
	if err != nil {
		return nil, err
	}

	if err := rds.ModifyUserRole(ctx, request.SpaceID, player_data.UID, request.Role); err != nil {
		return nil, err
	}

	players := rds.GetBattleSpacePlayers(ctx, request.SpaceID)
	notify := &mpubs.ModifyUserRoleNotify{
		SpaceID: request.SpaceID,
		Uid:     player_data.UID,
		Role:    request.Role,
	}

	for _, v := range players {
		actor.submitRequestGatewayPush(ctx, v, notify)
	}

	res := &mpubs.ModifyUserRoleResponse{
		Res: 0,
	}
	return res, nil
}

func (actor *BattleActor) onModifyUserCampRequest(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.ModifyUserCampRequest)
	sender := serve.GetServantClientInfo(ctx).Sender()

	player_data, err := rds.GetPlayerData(ctx, sender)
	if err != nil {
		return nil, err
	}
	if err := rds.ModifyUserCamp(ctx, request.SpaceID, player_data.UID, request.Camp); err != nil {
		return nil, err
	}

	players := rds.GetBattleSpacePlayers(ctx, request.SpaceID)
	notify := &mpubs.ModifyUserCampNotify{
		SpaceID: request.SpaceID,
		Uid:     player_data.UID,
		Camp:    request.Camp,
	}

	for _, v := range players {
		actor.submitRequestGatewayPush(ctx, v, notify)
	}

	res := &mpubs.ModifyUserCampResponse{
		Res: 0,
	}
	return res, nil
}

func (actor *BattleActor) onUserChatVoiceRequest(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.UserChatVoiceRequest)
	sender := serve.GetServantClientInfo(ctx).Sender()

	player_data, err := rds.GetPlayerData(ctx, sender)
	if err != nil {
		return nil, err
	}
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceID)
	notify := &mpubs.UserChatVoiceNotify{
		SpaceID: request.SpaceID,
		Uid:     player_data.UID,
		VoiceID: request.VoiceID,
	}

	for _, v := range players {
		actor.submitRequestGatewayPush(ctx, v, notify)
	}
	return nil, nil
}

func (actor *BattleActor) onCreateRobotRequest(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.CreateRobotRequest)
	sender := serve.GetServantClientInfo(ctx).Sender()
	if err := rds.IsMaster(ctx, sender); err != nil {
		return nil, err
	}
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceId)

	robotid, leaveuid, err := rds.CreateBattleSpaceRobot(ctx, request.SpaceId, request.Display, request.Icon, request.Pos, request.Role, request.Camp, request.Extends)
	if err != nil {
		return nil, err
	}

	if leaveuid != "" {
		notify := &mpubs.UserExitSpaceNotify{
			SpaceID: request.SpaceId,
			Uid:     leaveuid,
		}
		for _, v := range players {
			actor.submitRequestGatewayPush(ctx, v, notify)
		}
	}

	{
		player := &mpubs.BattleSpacePlayer{
			Uid:     robotid,
			Icon:    request.Icon,
			Display: request.Display,
			Pos:     request.Pos,
			Role:    request.Role,
			Camp:    request.Camp,
			Ready:   true,
			Extends: make(map[string]string),
			IsRobot: true,
		}

		for k, v := range request.Extends {
			player.Extends[k] = v
		}
		notify := &mpubs.EnterBattleSpaceNotify{
			SpaceId: request.SpaceId,
			Player:  player,
		}
		players = rds.GetBattleSpacePlayers(ctx, request.SpaceId)
		for _, v := range players {
			actor.submitRequestGatewayPush(ctx, v, notify)
		}
	}

	res := &mpubs.CreateRobotResponse{
		Res: 0,
	}
	return res, nil
}
func (actor *BattleActor) onRemoveRobotRequest(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.RemoveRobotRequest)
	sender := serve.GetServantClientInfo(ctx).Sender()

	if err := rds.IsMaster(ctx, sender); err != nil {
		return nil, err
	}
	if err := rds.RemoveBattleSpaceRobot(ctx, request.SpaceID, request.RobotID); err != nil {
		return nil, err
	}
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceID)

	notify := &mpubs.UserExitSpaceNotify{
		SpaceID: request.SpaceID,
		Uid:     request.RobotID,
	}

	for _, v := range players {
		actor.submitRequestGatewayPush(ctx, v, notify)
	}

	res := &mpubs.RemoveRobotResponse{
		Res: 0,
	}
	return res, nil
}
func (actor *BattleActor)onModifyRobotRoleRequest(ctx context.Context) (proto.Message, error){
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.ModifyRobotRoleRequest)
	sender := serve.GetServantClientInfo(ctx).Sender()
	if err := rds.IsMaster(ctx, sender); err != nil {
		return nil, err
	}
	if err := rds.ModifyUserRole(ctx, request.SpaceID, request.RobotID, request.Role); err != nil {
		return nil, err
	}
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceID)
	notify := &mpubs.ModifyUserRoleNotify{
		SpaceID: request.SpaceID,
		Uid:     request.RobotID,
		Role:    request.Role,
	}

	for _, v := range players {
		actor.submitRequestGatewayPush(ctx, v, notify)
	}

	res := &mpubs.ModifyRobotRoleResponse{
		Res: 0,
	}
	return res, nil

}
func(actor *BattleActor) onModifyRobotCampRequest(ctx context.Context) (proto.Message, error){
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.ModifyRobotCampRequest)
	sender := serve.GetServantClientInfo(ctx).Sender()
	if err := rds.IsMaster(ctx, sender); err != nil {
		return nil, err
	}
	if err := rds.ModifyUserCamp(ctx, request.SpaceID, request.RobotID, request.Camp); err != nil {
		return nil, err
	}
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceID)
	notify := &mpubs.ModifyUserCampNotify{
		SpaceID: request.SpaceID,
		Uid:     request.RobotID,
		Camp:    request.Camp,
	}

	for _, v := range players {
		actor.submitRequestGatewayPush(ctx, v, notify)
	}

	res := &mpubs.ModifyRobotCampResponse{
		Res: 0,
	}
	return res, nil
}

func (actor *BattleActor) onUserChatTextRequest(ctx context.Context) (proto.Message, error){
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.UserChatTextRequest)
	sender := serve.GetServantClientInfo(ctx).Sender()

	player_data, err := rds.GetPlayerData(ctx, sender)
	if err != nil {
		return nil, err
	}
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceID)
	notify := &mpubs.UserChatTextNotify{
		SpaceID: request.SpaceID,
		Uid:     player_data.UID,
		ChatText: request.ChatText,
	}

	for _, v := range players {
		actor.submitRequestGatewayPush(ctx, v, notify)
	}
	return nil, nil
}
func (actor *BattleActor) onDissBattleSpaceRequest(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mpubs.DissBattleSpaceRequest)
	sender := serve.GetServantClientInfo(ctx).Sender()
	vlog.Infof("onDissBattleSpaceRequest spaceid %s  ", request.SpaceId)
	// request := ctx.Message.(*mprvs.RequestExitBattleSpace)
	if err := rds.IsMaster(ctx, sender); err != nil {
		return nil, err
	}

	// 房主离开，解散房间
	players := rds.GetBattleSpacePlayers(ctx, request.SpaceId)
	if err := rds.DeleteBattleSpace(ctx, sender); err != nil {
		return nil, nil
	}

	notify := &mpubs.DissBattleSpaceNotify{
		SpaceId: request.SpaceId,
	}
	for _, v := range players {
		if !sender.Equal(v) {
			actor.submitRequestGatewayPush(ctx, v, notify)
		}
	}

	res := &mpubs.DissBattleSpaceResponse{
		SpaceId: request.SpaceId,
	}
	return res, nil
}

func (actor *BattleActor) onRequestExitBattleSpace(ctx context.Context) (proto.Message, error) {
	request := serve.GetServantClientInfo(ctx).Message().(*mprvs.RequestExitBattleSpace)
	sender := serve.GetServantClientInfo(ctx).Sender()
	// request := ctx.Message.(*mprvs.RequestExitBattleSpace)

	if err := rds.IsMaster(ctx, sender); err != nil {
		if err == errs.ErrorPermissionsLost {
			// 非房主离开，退出房间
			rds.LeaveBattleSpace(ctx, sender)
			res := &mpubs.UserExitSpaceNotify{
				SpaceID: request.BattleSpaceID,
				Uid:     request.UID,
			}
			players := rds.GetBattleSpacePlayers(ctx, request.BattleSpaceID)
			for _, v := range players {
				if !sender.Equal(v) {
					actor.submitRequestGatewayPush(ctx, v, res)
				}
			}
		} else {
			return nil, err
		}
	} else {
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
