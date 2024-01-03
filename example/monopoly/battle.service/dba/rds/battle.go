package rds

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lithammer/shortuuid/v4"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/errs"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
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
	uid string,
	clientId *network.ClientID,
	mapURi string) (string, error) {

	mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return "", err
	}

	results, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientIcon,
		rdsconst.PlayerMapClientBattleSpaceId).Result()
	if err != nil {
		mutex.Unlock()
		return "", err
	}

	if len(results) != 2 {
		mutex.Unlock()
		return "", errs.ErrorPlayerOnlineDataLost
	}

	if results[1] != "" {
		mutex.Unlock()
		return "", errs.ErrorPlayerAlreadyInBattleSpace
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Do(ctx, "MULTI")

	results[1] = shortuuid.New()

	battleSpace := map[string]string{
		rdsconst.BattleSpaceId:            results[1].(string),
		rdsconst.PalyerMapClientIdAddress: clientId.Address,
		rdsconst.PlayerMapClientIdId:      clientId.Id,
		rdsconst.BattleSpaceMapURi:        mapURi,
		rdsconst.BattleSpaceNatAddr:       "",
		rdsconst.BattleSpaceMasterUid:     uid,
		rdsconst.BattleSpaceMasterIcon:    results[0].(string),
		rdsconst.BattleSpaceMasterDisplay: rdsconst.BattleSpaceStateReady,
		rdsconst.BattleSpacePos1Uid:       uid,
		rdsconst.BattleSpacePos2Uid:       "",
		rdsconst.BattleSpacePos3Uid:       "",
		rdsconst.BattleSpacePos4Uid:       "",

		rdsconst.BattleSpacePos1Icon: results[0].(string),
		rdsconst.BattleSpacePos2Icon: results[0].(string),
		rdsconst.BattleSpacePos3Icon: results[0].(string),
		rdsconst.BattleSpacePos4Icon: results[0].(string),

		rdsconst.BattleSpacePos1Display: rdsconst.BattleSpaceStateReady,
		rdsconst.BattleSpacePos2Display: "",
		rdsconst.BattleSpacePos3Display: "",
		rdsconst.BattleSpacePos4Display: "",

		rdsconst.BattleSpaceTime:  strconv.FormatInt(time.Now().UnixMilli(), 10),
		rdsconst.BattleSpaceState: rdsconst.BattleSpaceStateNomal,
	}

	pipe.HSet(ctx, rdsconst.GetPlayerOnlineDataKey(uid), rdsconst.PlayerMapClientBattleSpaceId, results[1])
	pipe.HMSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(results[1].(string)), battleSpace)
	pipe.RPush(ctx, rdsconst.BattleSpaceOnlinetable, results[1].(string))

	pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		mutex.Unlock()
		return "", err
	}

	mutex.Unlock()

	return results[0].(string), nil
}

func DeleteBattleSpace(ctx context.Context, spaceid string, clientId *network.ClientID) error {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}

		return err
	}

	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := player_mutex.Lock(); err != nil {
		return err
	}
	results, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientBattleSpaceId).Result()
	if err != nil {
		player_mutex.Unlock()
		return err
	}
	if len(results) != 1 {
		player_mutex.Unlock()
		return errs.ErrorPlayerOnlineDataLost
	}
	if results[0] != spaceid {
		player_mutex.Unlock()
		return errs.ErrorPlayerIsNotInBattleSpace
	}
	oldSpaceid := results[0].(string)
	if oldSpaceid != spaceid {
		player_mutex.Unlock()
		return errs.ErrorPlayerIsNotInBattleSpace
	}

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceOnlineDataKey(oldSpaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}

	spaceResults, err := client.HGetAll(ctx, rdsconst.GetBattleSpaceOnlineDataKey(oldSpaceid)).Result()
	if err != nil {
		space_mutex.Unlock()
		player_mutex.Unlock()
		return err
	}

	if !clientId.Equal(&network.ClientID{Address: spaceResults[rdsconst.PalyerMapClientIdAddress],
		Id: spaceResults[rdsconst.PlayerMapClientIdId]}) {
		space_mutex.Unlock()
		player_mutex.Unlock()
		return errors.New("permissions lost")
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Do(ctx, "MULTI")

	pipe.HSet(ctx, rdsconst.GetPlayerOnlineDataKey(uid), rdsconst.PlayerMapClientBattleSpaceId, "")
	pipe.HDel(ctx, rdsconst.GetBattleSpaceOnlineDataKey(oldSpaceid))
	pipe.LRem(ctx, rdsconst.BattleSpaceOnlinetable, 1, oldSpaceid)
	pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		space_mutex.Unlock()
		player_mutex.Unlock()
		return err
	}

	space_mutex.Unlock()
	player_mutex.Unlock()
	return nil
}

func EnterBattleSpace(ctx context.Context, spaceid string, clientId *network.ClientID) error {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return  nil
		}

		return err
	}

	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := player_mutex.Lock(); err != nil {
		return err
	}
	playerResults, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientBattleSpaceId).Result()
	if err != nil {
		player_mutex.Unlock()
		return  err
	}
	if len(playerResults) != 1 {
		player_mutex.Unlock()
		return  errs.ErrorPlayerOnlineDataLost
	}
	if playerResults[0] != "" {
		player_mutex.Unlock()
		return errs.ErrorPlayerAlreadyInBattleSpace
	}

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceOnlineDataKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		player_mutex.Unlock()
		return err
	}

	spaceResults, err := client.HGetAll(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		space_mutex.Unlock()
		player_mutex.Unlock()
		return err
	}
	if spaceResults[rdsconst.BattleSpacePos1Uid] != "" &&
		spaceResults[rdsconst.BattleSpacePos2Uid] != "" &&
		spaceResults[rdsconst.BattleSpacePos3Uid] != "" &&
		spaceResults[rdsconst.BattleSpacePos4Uid] != "" {
		return errs.ErrorSpacePlayerIsFull
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Do(ctx, "MULTI")
	pipe.HSet(ctx, rdsconst.GetPlayerOnlineDataKey(uid), rdsconst.PlayerMapClientBattleSpaceId, spaceid)
	if spaceResults[rdsconst.BattleSpacePos1Uid] == "" {
		pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos1Uid, uid)
		// pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos1Icon, playerResults[rdsconst.p])
	} else if spaceResults[rdsconst.BattleSpacePos2Uid] == "" {
		pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos2Uid, uid)
	} else if spaceResults[rdsconst.BattleSpacePos3Uid] == "" {
		pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos3Uid, uid)
	} else if spaceResults[rdsconst.BattleSpacePos4Uid] == "" {
		pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos4Uid, uid)
	}

	pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		space_mutex.Unlock()
		player_mutex.Unlock()
		return err
	}

	space_mutex.Unlock()
	player_mutex.Unlock()

	return nil
}

func DisplayBattleSpace(ctx context.Context, spaceid string, display string, clientId *network.ClientID) error {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := player_mutex.Lock(); err != nil {
		return err
	}
	playerResults, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientBattleSpaceId).Result()
	if err != nil {
		player_mutex.Unlock()
		return err
	}
	if len(playerResults) != 1 {
		player_mutex.Unlock()
		return errs.ErrorPlayerOnlineDataLost
	}
	if playerResults[0] != spaceid {
		player_mutex.Unlock()
		return errs.ErrorPlayerIsNotInBattleSpace
	}
	player_mutex.Unlock()

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceOnlineDataKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}

	results, err := client.HGetAll(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		space_mutex.Unlock()
		return err
	}
	if results[rdsconst.BattleSpacePos1Uid] != uid &&
		results[rdsconst.BattleSpacePos2Uid] != uid &&
		results[rdsconst.BattleSpacePos3Uid] != uid &&
		results[rdsconst.BattleSpacePos4Uid] != uid {
		return  errs.ErrorPlayerIsNotInBattleSpace
	}
	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Do(ctx, "MULTI")
	if results[rdsconst.BattleSpacePos1Uid] == uid {
		pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos1Display, display)
		// pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos1Icon, playerResults[rdsconst.p])
	} else if results[rdsconst.BattleSpacePos2Uid] == uid {
		pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos2Display, display)
	} else if results[rdsconst.BattleSpacePos3Uid] == uid {
		pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos3Display, display)
	} else if results[rdsconst.BattleSpacePos4Uid] == uid {
		pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos4Display, display)
	}

	pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		space_mutex.Unlock()
		return  err
	}
	space_mutex.Unlock()
	return nil
}

func LeaveBattleSpace(ctx context.Context, spaceid string, clientId *network.ClientID) error {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := player_mutex.Lock(); err != nil {
		return err
	}
	playerResults, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientBattleSpaceId).Result()
	if err != nil {
		player_mutex.Unlock()
		return err
	}
	if len(playerResults) != 1 {
		player_mutex.Unlock()
		return errs.ErrorPlayerOnlineDataLost
	}
	if playerResults[0] != spaceid {
		player_mutex.Unlock()
		return errs.ErrorPlayerIsNotInBattleSpace
	}
	player_mutex.Unlock()
	
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceOnlineDataKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}

	results, err := client.HGetAll(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		space_mutex.Unlock()
		return err
	}

	

	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Do(ctx, "MULTI")
	//TODO 房主退出
	if clientId.Equal(&network.ClientID{Address: results[rdsconst.PalyerMapClientIdAddress],
		Id: results[rdsconst.PlayerMapClientIdId]}) {

	}else{
		if results[rdsconst.BattleSpacePos1Uid] == uid {
			pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos1Display, rdsconst.BattleSpaceStateNomal)
			pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid),rdsconst.BattleSpacePos1Uid, 0)
		} else if results[rdsconst.BattleSpacePos2Uid] == uid {
			pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos2Display, rdsconst.BattleSpaceStateNomal)
			pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid),rdsconst.BattleSpacePos2Uid, 0)
		} else if results[rdsconst.BattleSpacePos3Uid] == uid {
			pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos3Display, rdsconst.BattleSpaceStateNomal)
			pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid),rdsconst.BattleSpacePos3Uid, 0)
		} else if results[rdsconst.BattleSpacePos4Uid] == uid {
			pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePos4Display, rdsconst.BattleSpaceStateNomal)
			pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid),rdsconst.BattleSpacePos4Uid, 0)
		}
	}
	pipe.HSet(ctx, rdsconst.GetPlayerOnlineDataKey(uid), rdsconst.PlayerMapClientBattleSpaceId, "")
	//TODO 玩家退出
	pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		space_mutex.Unlock()
		return  err
	}
	space_mutex.Unlock()
	return nil
}

func GetBattleSpaceList(ctx context.Context, start int64, size int64) ([]string, error) {

	val, err := client.LRange(ctx, rdsconst.GetBattleSpaceOnlineDataKey(""), start, (start+1)*size).Result()
	if err != nil {
		return nil, err
	}

	return val, nil
}

func AutoEnterBattleSpace(ctx context.Context, clientId *network.ClientID) error {
	return nil
}

func GetBattleSpacesCount(ctx context.Context) (int64, error) {
	return client.LLen(ctx, rdsconst.GetBattleSpaceOnlineDataKey("")).Result()
}

func FindBattleSpaceData(ctx context.Context, spaceid string) (map[string]string, error) {
	mutex := sync.NewMutex(rdsconst.GetBattleSpaceOnlineDataKey(spaceid))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	results, err := client.HGetAll(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		mutex.Unlock()
		return nil, err
	}
	mutex.Unlock()

	return results, nil
}
