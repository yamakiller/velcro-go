package rds

import (
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lithammer/shortuuid/v4"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/errs"
	mrdsstruct "github.com/yamakiller/velcro-go/example/monopoly/protocols/rdsstruct"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsstruct"
	"github.com/yamakiller/velcro-go/network"
)

//1.创建对战区 对战区长时不开始且没有任何人加入，将在10分钟后结束
//2.删除对战区
//3.加入对战区
//4.离开对战区
//5.查询对战区列表
//6.自动加入某对战区
//7.对战始流传: 7-1.通知主机提交UDP(既:Nat信息)
//				 7-2.收到Nat信息回复客户端，已收到.(不需要再发送)
//				 7-3.将Nat信息发送给参与者
//				 7-4.通知各方开始游戏:
//									7-4-1.如果通知失败,通知主机对战取消,对战区取消开始状态
//									7-4-2.修改对战区状态,修改对战区倒计时时间

func CreateBattleSpace(ctx context.Context,
	master *rdsstruct.RdsPlayerData,
	clientId *network.ClientID,
	mapURi string,
	max_count int32,
	spaceName string,
	password string,
	extend string,
	display string) (string, error) {

	pipe := client.TxPipeline()
	defer pipe.Close()

	// pipe.Do(ctx, "MULTI")

	spaceid := shortuuid.New()

	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	battleSpace.SpaceId = spaceid
	battleSpace.SpaceName = spaceName
	battleSpace.SpacePassword = password
	battleSpace.SpaceExtend = extend
	battleSpace.SpaceMapURI = mapURi
	battleSpace.SpaceMasterUid = master.UID
	battleSpace.SpaceMasterClientAddress = master.ClientIdAddress
	battleSpace.SpaceMasterClinetID = master.ClientIdId
	battleSpace.SpaceMasterDisplay = display
	battleSpace.SpaceMasterIcon = master.Externs[rdsconst.PlayerMapClientIcon]
	battleSpace.SpaceStarttime = time.Now().UnixMilli()
	battleSpace.SpaceState = rdsconst.BattleSpaceStateNomal
	battleSpace.SpacePlayers = make([]*mrdsstruct.RdsBattleSpacePlayer, max_count)
	battleSpace.SpacePlayers[0] = &mrdsstruct.RdsBattleSpacePlayer{
		Uid:     master.UID,
		Display: display,
		Icon:    master.Externs[rdsconst.PlayerMapClientIcon],
		Pos:     0,
		Extends: make(map[string]string),
	}

	for k, v := range master.Externs {
		battleSpace.SpacePlayers[0].Extends[k] = v
	}

	BattleSpaceDieTime := time.Duration(envs.Instance().Get("configs").(*configs.Config).Server.BattleSpaceDieTime) * time.Second
	pipe.Set(ctx, rdsconst.GetPlayerBattleSpaceIDKey(master.UID), spaceid, BattleSpaceDieTime)
	pipe.Set(ctx, rdsconst.GetPlayerBattleSpaceIDAndUIDKey(master.UID, spaceid), spaceid, BattleSpaceDieTime)
	pipe.Set(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), battleSpace, BattleSpaceDieTime)
	if BattleSpaceDieTime > 30*time.Second {
		pipe.Set(ctx, rdsconst.GetBattleSpaceOnlineExpiredKey(spaceid), spaceid, BattleSpaceDieTime-30*time.Second)
	}
	pipe.RPush(ctx, rdsconst.BattleSpaceOnlinetable, spaceid)

	// pipe.Do(ctx, "exec")

	_, err := pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return "", err
	}

	return spaceid, nil
}

func DeleteBattleSpace(ctx context.Context, clientId *network.ClientID) error {

	player_data, err := GetPlayerData(ctx, clientId)
	if err != nil {
		return nil
	}
	BattleSpaceId, err := GetPlayerBattleSpaceID(ctx, player_data.UID)
	if err != nil || BattleSpaceId == "" {
		return errs.ErrorPlayerIsNotInBattleSpace
	}

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(BattleSpaceId))
	if err = space_mutex.Lock(); err != nil {
		return err
	}
	defer space_mutex.Unlock()
	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(BattleSpaceId)).Scan(battleSpace); err != nil {
		return err
	}

	pipe := client.TxPipeline()
	defer pipe.Close()
	pipe.Del(ctx, rdsconst.GetBattleSpaceOnlineDataKey(BattleSpaceId))
	pipe.Del(ctx, rdsconst.GetBattleSpaceOnlineExpiredKey(BattleSpaceId))
	pipe.LRem(ctx, rdsconst.BattleSpaceOnlinetable, 1, BattleSpaceId)
	// pipe.Do(ctx, "MULTI")
	for _, v := range battleSpace.SpacePlayers {
		if v.Uid != "" {
			pipe.Del(ctx, rdsconst.GetPlayerBattleSpaceIDKey(v.Uid))
			pipe.Del(ctx, rdsconst.GetPlayerBattleSpaceIDAndUIDKey(v.Uid, BattleSpaceId))

		}
	}

	// pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}
	return nil
}

func EnterBattleSpace(ctx context.Context, spaceid string, password string, display string, clientId *network.ClientID) error {
	player_data, err := GetPlayerData(ctx, clientId)
	if err != nil {
		return err
	}

	BattleSpaceId, _ := GetPlayerBattleSpaceID(ctx, player_data.UID)

	if BattleSpaceId != "" && BattleSpaceId != spaceid {
		return errs.ErrorPlayerAlreadyInBattleSpace
	}

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}
	defer space_mutex.Unlock()

	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Scan(battleSpace); err != nil {
		return err
	}
	for i, v := range battleSpace.SpacePlayers {
		if v.Uid == player_data.UID {
			battleSpace.SpacePlayers[i] = &mrdsstruct.RdsBattleSpacePlayer{
				Uid:     player_data.UID,
				Display: display,
				Icon:    player_data.Externs[rdsconst.PlayerMapClientIcon],
				Pos:     int32(i),
				Extends: make(map[string]string),
			}
			for k, v := range player_data.Externs {
				battleSpace.SpacePlayers[i].Extends[k] = v
			}
			goto enter_lable
		}
	}
	if battleSpace.SpacePassword != password {
		return errs.ErrorSpacePassword
	}
	{
		isFull := true
		for i, v := range battleSpace.SpacePlayers {
			if v.Uid == "" {
				battleSpace.SpacePlayers[i] = &mrdsstruct.RdsBattleSpacePlayer{
					Uid:     player_data.UID,
					Display: display,
					Icon:    player_data.Externs[rdsconst.PlayerMapClientIcon],
					Pos:     int32(i),
					Extends: make(map[string]string),
				}
				for k, v := range player_data.Externs {
					battleSpace.SpacePlayers[i].Extends[k] = v
				}
				isFull = false
				break
			}
		}

		if isFull {
			return errs.ErrorSpacePlayerIsFull
		}
	}
enter_lable:
	expire, err := client.TTL(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		return nil
	}
	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Set(ctx, rdsconst.GetPlayerBattleSpaceIDKey(player_data.UID), spaceid, expire)
	pipe.Set(ctx, rdsconst.GetPlayerBattleSpaceIDAndUIDKey(player_data.UID, spaceid), spaceid, expire)
	pipe.Set(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), battleSpace, expire)

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}
	return nil
}

func ChangeBattleSpacePassword(ctx context.Context, spaceid string, new_password string) error {
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}
	defer space_mutex.Unlock()
	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Scan(battleSpace); err != nil {
		return err
	}
	battleSpace.SpacePassword = new_password

	expire, err := client.TTL(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		return nil
	}

	pipe := client.TxPipeline()
	defer pipe.Close()
	pipe.Set(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), battleSpace, expire)
	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}
	return nil
}

func ChangeModifyRoomParameters(ctx context.Context, spaceid string, map_url string, max_count uint32, room_name string, extend string) error {
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}
	defer space_mutex.Unlock()

	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Scan(battleSpace); err != nil {
		return err
	}

	if map_url != "" {
		battleSpace.SpaceMapURI = map_url
	}
	if room_name != "" {
		battleSpace.SpaceName = room_name
	}
	if extend != "" {
		battleSpace.SpaceExtend = extend
	}
	expire, err := client.TTL(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		return nil
	}
	pipe := client.TxPipeline()
	defer pipe.Close()
	if max_count > 1 {
		space_pos := make([]*mrdsstruct.RdsBattleSpacePlayer, max_count)
		for i, v := range battleSpace.SpacePlayers {
			if i >= len(space_pos) {
				if v.Uid != "" {
					pipe.Del(ctx, rdsconst.GetPlayerBattleSpaceIDKey(v.Uid))
					pipe.Del(ctx, rdsconst.GetPlayerBattleSpaceIDAndUIDKey(v.Uid, spaceid))

				}
				continue
			}
			space_pos[i] = v
		}
		battleSpace.SpacePlayers = space_pos
	}
	pipe.Set(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), battleSpace, expire)

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}
	return nil
}

func ModifyUserRole(ctx context.Context, spaceid string, uid string, role string) error {
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}
	defer space_mutex.Unlock()
	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Scan(battleSpace); err != nil {
		return err
	}

	for _, v := range battleSpace.SpacePlayers {
		if v.Uid == uid {
			v.Role = role
			break
		}
	}

	expire, err := client.TTL(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		return nil
	}
	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Set(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), battleSpace, expire)
	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}
	return nil
}

func ModifyUserCamp(ctx context.Context, spaceid string, uid string, camp string) error {
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}
	defer space_mutex.Unlock()
	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Scan(battleSpace); err != nil {
		return err
	}

	for _, v := range battleSpace.SpacePlayers {
		if v.Uid == uid {
			v.Camp = camp
			break
		}
	}

	expire, err := client.TTL(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		return nil
	}
	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Set(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), battleSpace, expire)
	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}
	return nil
}

func ReadyBattleSpace(ctx context.Context, spaceid string, ready bool, clientId *network.ClientID) error {
	player_data, err := GetPlayerData(ctx, clientId)
	if err != nil {
		return nil
	}

	BattleSpaceId, _ := GetPlayerBattleSpaceID(ctx, player_data.UID)
	if BattleSpaceId != spaceid {
		return errs.ErrorPlayerIsNotInBattleSpace
	}

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {

		return err
	}
	defer space_mutex.Unlock()

	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Scan(battleSpace); err != nil {
		return err
	}

	for _, v := range battleSpace.SpacePlayers {
		if v.Uid == player_data.UID {
			v.Ready = ready
			break
		}
	}
	expire, err := client.TTL(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		return nil
	}
	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Set(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), battleSpace, expire)
	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}
	return nil
}

func StartBattleSpace(ctx context.Context, spaceid string, clientId *network.ClientID) error {
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}
	defer space_mutex.Unlock()

	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Scan(battleSpace); err != nil {
		return err
	}

	if !clientId.Equal(&network.ClientID{Address: battleSpace.SpaceMasterClientAddress,
		Id: battleSpace.SpaceMasterClinetID}) {
		return errs.ErrorPermissionsLost
	}

	if battleSpace.SpaceState == rdsconst.BattleSpaceStateRunning {
		return errs.ErrorSpaceIsRunning
	}

	battleSpace.SpaceState = rdsconst.BattleSpaceStateRunning

	expire, err := client.TTL(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		return nil
	}
	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Set(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), battleSpace, expire)
	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}
	return nil
}

func GetBattleSpaceInfo(ctx context.Context, spaceid string) (*rdsstruct.RdsBattleSpaceData, error) {
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return nil, err
	}
	defer space_mutex.Unlock()

	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Scan(battleSpace); err != nil {
		return nil, err
	}

	return battleSpace, nil
}
func LeaveBattleSpace(ctx context.Context, clientId *network.ClientID) {
	player_data, err := GetPlayerData(ctx, clientId)
	if err != nil {
		return
	}
	BattleSpaceId, _ := GetPlayerBattleSpaceID(ctx, player_data.UID)

	if BattleSpaceId == "" {
		return
	}

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(BattleSpaceId))
	if err = space_mutex.Lock(); err != nil {
		return
	}
	defer space_mutex.Unlock()

	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(BattleSpaceId)).Scan(battleSpace); err != nil {
		return
	}

	for i, v := range battleSpace.SpacePlayers {
		if v.Uid == player_data.UID {
			battleSpace.SpacePlayers[i] = nil
			break
		}
	}
	expire, err := client.TTL(ctx, rdsconst.GetBattleSpaceOnlineDataKey(BattleSpaceId)).Result()
	if err != nil {
		return
	}
	pipe := client.TxPipeline()
	defer pipe.Close()
	pipe.Del(ctx, rdsconst.GetPlayerBattleSpaceIDKey(player_data.UID))
	pipe.Del(ctx, rdsconst.GetPlayerBattleSpaceIDAndUIDKey(player_data.UID, BattleSpaceId))
	pipe.Set(ctx, rdsconst.GetBattleSpaceOnlineDataKey(BattleSpaceId), battleSpace, expire)
	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()

		return
	}

	return
}

func GetBattleSpaceList(ctx context.Context, start int64, size int64) ([]string, error) {
	//pipe.RPush(ctx, rdsconst.BattleSpaceOnlinetable, spaceid)
	//val, err := client.LRange(ctx, rdsconst.GetBattleSpaceOnlineDataKey(""), start, (start+1)*size).Result()
	val, err := client.LRange(ctx, rdsconst.BattleSpaceOnlinetable, start*size, (start+1)*size).Result()
	if err != nil {
		return nil, err
	}

	return val, nil
}

func AutoEnterBattleSpace(ctx context.Context, clientId *network.ClientID) error {
	return nil
}

func GetBattleSpacesCount(ctx context.Context) (int64, error) {
	return client.LLen(ctx, rdsconst.BattleSpaceOnlinetable).Result()
}

func GetBattleSpacePlayers(ctx context.Context, spaceid string) []*network.ClientID {
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return nil
	}

	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Scan(battleSpace); err != nil {
		space_mutex.Unlock()
		return nil
	}
	space_mutex.Unlock()

	list := make([]*network.ClientID, 0)
	for _, v := range battleSpace.SpacePlayers {
		if v.Uid == "" {
			continue
		}
		cli := GetBattleSpacePlayerClientID(ctx, v.Uid)
		if cli == nil {
			continue
		}
		list = append(list, cli)
	}

	return list
}

func IsMaster(ctx context.Context, clientId *network.ClientID) (bool, error) {
	player_data, err := GetPlayerData(ctx, clientId)
	if err != nil {
		return false, err
	}
	BattleSpaceId, err := GetPlayerBattleSpaceID(ctx, player_data.UID)
	if BattleSpaceId == "" || err != nil {
		return false, errs.ErrorPlayerIsNotInBattleSpace
	}

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(BattleSpaceId))
	if err := space_mutex.Lock(); err != nil {
		return false, err
	}

	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(BattleSpaceId)).Scan(battleSpace); err != nil {
		space_mutex.Unlock()
		return false, err
	}
	space_mutex.Unlock()

	return battleSpace.SpaceMasterUid == player_data.UID, nil
}

func IsBattleSpaceMaster(ctx context.Context, clientId *network.ClientID, spaceid string) error {
	player_data, err := GetPlayerData(ctx, clientId)
	if err != nil {
		return err
	}

	BattleSpaceId, err := GetPlayerBattleSpaceID(ctx, player_data.UID)
	if BattleSpaceId == "" || err != nil {
		return errs.ErrorPlayerIsNotInBattleSpace
	}

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}

	battleSpace := &rdsstruct.RdsBattleSpaceData{}
	if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(BattleSpaceId)).Scan(battleSpace); err != nil {
		space_mutex.Unlock()
		return err
	}
	space_mutex.Unlock()

	if battleSpace.SpaceMasterUid == player_data.UID {
		return nil
	}

	return errs.ErrorPlayerIsNotInBattleSpace
}

func ClearBattleSpace(ctx context.Context) {
	spaceids, err := GetBattleSpaceList(ctx, 0, 1000000)
	if err != nil {
		return
	}
	pipe := client.TxPipeline()
	defer pipe.Close()
	for _, spaceid := range spaceids {
		battleSpace := &rdsstruct.RdsBattleSpaceData{}
		if err := client.Get(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Scan(battleSpace); err == nil && battleSpace != nil {
			for _, v := range battleSpace.SpacePlayers {
				if v.Uid != "" {
					pipe.Del(ctx, rdsconst.GetPlayerBattleSpaceIDKey(v.Uid))
					pipe.Del(ctx, rdsconst.GetPlayerBattleSpaceIDAndUIDKey(v.Uid, spaceid))

				}
			}
		}

		pipe.LRem(ctx, rdsconst.BattleSpaceOnlinetable, 1, spaceid)
		pipe.Del(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid))
		pipe.Del(ctx, rdsconst.GetBattleSpaceOnlineExpiredKey(spaceid))

	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
	}
}

func RemoveBattleSpaceList(spaceid string) {
	client.LRem(context.Background(), rdsconst.BattleSpaceOnlinetable, 1, spaceid)
}

func SubscribeBattleSpaceTime(ctx context.Context) *redis.PubSub {
	// 开启检测
	_, err := client.ConfigSet(context.Background(), "notify-keyspace-events", "Ex").Result()
	if err != nil {
		panic(err)
	}
	pubsub := client.Subscribe(ctx, "__keyevent@0__:expired")

	// 开启goroutine，接收过期事件
	go func() {
		for msg := range pubsub.Channel() {
			// 处理过期事件
			// 警告过期时间
			// 删除过期房间号
			// 通知房间解散
			if strings.HasPrefix(msg.Payload, rdsconst.GetBattleSpaceOnlineExpiredKey("")) {
				list := rdsconst.GetBattleSpaceOnlineExpiredSpaceID(msg.Payload)
				if len(list) == 2 {
					sendDisRoomWarningNotify(list[1], 30)
				}
			} else if strings.HasPrefix(msg.Payload, rdsconst.GetBattleSpaceOnlineDataKey("")) {
				list := rdsconst.GetBattleSpaceOnlineDataSpaceID(msg.Payload)
				if len(list) == 2 {
					RemoveBattleSpaceList(list[1])
				}
			} else if strings.HasPrefix(msg.Payload, rdsconst.PlayerBattleSpaceIDUID) {
				list := rdsconst.GetPlayerBattleSpaceIDAndUID(msg.Payload)
				if len(list) == 3 {
					cli := GetBattleSpacePlayerClientID(context.Background(), list[1])
					if cli != nil {
						sendDissBattleSpaceNotify(list[2], cli)
					}
				}
			}
		}
	}()
	return pubsub
}
