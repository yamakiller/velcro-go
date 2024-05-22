package rds

import (
	"context"
	"errors"
	"fmt"
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
	master string,
	clientId *network.ClientID,
	mapURi string,
	max_count int32) (string, error) {

	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(master))
	if err := player_mutex.Lock(); err != nil {
		return "", err
	}
	results, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(master),
		rdsconst.PlayerMapClientIcon,
		rdsconst.PlayerMapClientBattleSpaceId,
		rdsconst.PlayerMapClientDisplayName).Result()
	if err != nil {
		player_mutex.Unlock()
		return "", err
	}
	player_mutex.Unlock()
	if len(results) != 3 {
		return "", errs.ErrorPlayerOnlineDataLost
	}

	if results[1] != nil && results[1].(string) != "" {
		return "", errs.ErrorPlayerAlreadyInBattleSpace
	}

	player_icon := ""
	if results[0] != nil {
		player_icon = results[0].(string)
	}

	player_display := ""
	if results[2] != nil {
		player_display = results[2].(string)
	}
	pipe := client.TxPipeline()
	defer pipe.Close()

	// pipe.Do(ctx, "MULTI")

	spaceid := shortuuid.New()

	battleSpace := map[string]string{
		rdsconst.BattleSpaceId:                    spaceid,
		rdsconst.PalyerMapClientIdAddress:         clientId.Address,
		rdsconst.PlayerMapClientIdId:              clientId.Id,
		rdsconst.BattleSpaceMapURi:                mapURi,
		rdsconst.BattleSpaceNatAddr:               "",
		rdsconst.BattleSpaceMasterUid:             master,
		rdsconst.BattleSpaceMasterIcon:            player_icon,
		rdsconst.BattleSpaceMasterDisplay:         rdsconst.BattleSpaceStateReady,
		rdsconst.BattleSpacePlayerCount:           strconv.FormatInt(int64(max_count), 10),
		rdsconst.BattleSpacePlayerPos:             rdsconst.UpdateData(make([]string, max_count), 0, master),
		rdsconst.GetBattleSpacePlayerDataKey(master): fmt.Sprintf("%s&%s&%s&%s&%s", master, player_icon, player_display, rdsconst.BattleSpaceStateReady, "0"),
		rdsconst.BattleSpaceTime:                  strconv.FormatInt(time.Now().UnixMilli(), 10),
		rdsconst.BattleSpaceState:                 rdsconst.BattleSpaceStateNomal,
	}

	pipe.HSet(ctx, rdsconst.GetPlayerOnlineDataKey(master), rdsconst.PlayerMapClientBattleSpaceId, spaceid)
	pipe.HMSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), battleSpace)
	pipe.RPush(ctx, rdsconst.BattleSpaceOnlinetable, spaceid)

	// pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return "", err
	}

	return spaceid, nil
}

func DeleteBattleSpace(ctx context.Context, clientId *network.ClientID) (err error) {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		return
	}

	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err = player_mutex.Lock(); err != nil {
		return
	}

	results, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientBattleSpaceId).Result()
	if err != nil {
		player_mutex.Unlock()
		return
	}
	player_mutex.Unlock()

	if len(results) != 1 {
		err = errs.ErrorPlayerOnlineDataLost
		return
	}
	if results[0] == nil || results[0].(string) == "" {
		err = errs.ErrorPlayerIsNotInBattleSpace
		return
	}

	spaceid := results[0].(string)

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err = space_mutex.Lock(); err != nil {
		return
	}
	defer space_mutex.Unlock()
	spaceResults, err := client.HGetAll(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		return
	}

	if spaceResults[rdsconst.BattleSpaceMasterUid] != uid {
		err = errs.ErrorPermissionsLost
		return
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	// pipe.Do(ctx, "MULTI")

	pipe.HMSet(ctx, rdsconst.GetPlayerOnlineDataKey(uid), rdsconst.PlayerMapClientBattleSpaceId, "")
	pipe.Del(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid))
	pipe.LRem(ctx, rdsconst.BattleSpaceOnlinetable, 1, spaceid)
	// pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return
	}
	return
}

func EnterBattleSpace(ctx context.Context, spaceid string, clientId *network.ClientID) error {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		return err
	}

	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := player_mutex.Lock(); err != nil {
		return err
	}

	playerResults, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientIcon,
		rdsconst.PlayerMapClientBattleSpaceId,
		rdsconst.PlayerMapClientDisplayName).Result()
	if err != nil {
		player_mutex.Unlock()
		return err
	}
	player_mutex.Unlock()

	if len(playerResults) != 3 {
		return errs.ErrorPlayerOnlineDataLost
	}
	if playerResults[1] != nil && playerResults[1].(string) != "" {
		return errs.ErrorPlayerAlreadyInBattleSpace
	}

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}
	defer space_mutex.Unlock()
	pos_result, err := client.HMGet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePlayerPos).Result()
	if err != nil {
		return err
	}

	max_player_count, err := client.HMGet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePlayerCount).Result()
	if err != nil {

		return err
	}

	players := rdsconst.SplitData(pos_result[0].(string))
	count := 0
	for _, id := range players {
		if id == uid {
			return errs.ErrorPlayerAlreadyInBattleSpace
		}
		if id != "" {
			count++
		}
	}

	max_count, _ := strconv.Atoi(max_player_count[0].(string))

	if max_count <= count {
		return errs.ErrorSpacePlayerIsFull
	}

	player_icon := ""
	if playerResults[0] != nil {
		player_icon = playerResults[0].(string)
	}

	player_display := ""
	if playerResults[2] != nil {
		player_display = playerResults[2].(string)
	}

	space_pos, index := rdsconst.EnterData(rdsconst.SplitData(pos_result[0].(string)), uid)
	pipe := client.TxPipeline()
	defer pipe.Close()

	// pipe.Do(ctx, "MULTI")
	pipe.HMSet(ctx, rdsconst.GetPlayerOnlineDataKey(uid), rdsconst.PlayerMapClientBattleSpaceId, spaceid)
	pipe.HMSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePlayerPos, space_pos)
	pipe.HMSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid),
		rdsconst.GetBattleSpacePlayerDataKey(uid),
		fmt.Sprintf("%s&%s&%s&%s&%s", uid, player_icon, player_display, rdsconst.BattleSpaceStateReady, strconv.FormatInt(int64(index), 10)),
	)

	// pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}
	return nil
}

func ReadyBattleSpace(ctx context.Context, spaceid string, ready bool, clientId *network.ClientID) error {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
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
	player_mutex.Unlock()
	if len(playerResults) != 1 {
		return errs.ErrorPlayerOnlineDataLost
	}
	if playerResults[0] != spaceid {
		return errs.ErrorPlayerIsNotInBattleSpace
	}

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {

		return err
	}
	defer space_mutex.Unlock()
	results, err := client.HMGet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.GetBattleSpacePlayerDataKey(uid)).Result()
	if err != nil {

		return err
	}

	if len(results) != 1 {
		return errs.ErrorPlayerIsNotInBattleSpace
	}
	player_data := rdsconst.SplitData(results[0].(string))
	if ready {
		player_data[3] = rdsconst.BattleSpaceStateReady
	} else {
		player_data[3] = rdsconst.BattleSpaceStateNomal
	}
	// player_data := makeBattleSpacePlayer(id,icon,display,read,pos)
	pipe := client.TxPipeline()
	defer pipe.Close()

	// pipe.Do(ctx, "MULTI")
	pipe.HMSet(ctx,
		rdsconst.GetBattleSpaceOnlineDataKey(spaceid),
		rdsconst.GetBattleSpacePlayerDataKey(uid),
		fmt.Sprintf("%s&%s&%s&%s&%s", player_data[0], player_data[1], player_data[2], player_data[3], player_data[4]),
	)

	// pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}
	return nil
}

func StartBattleSpace(ctx context.Context, spaceid string, clientId *network.ClientID) (err error) {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		return
	}

	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err = player_mutex.Lock(); err != nil {
		return
	}

	results, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientBattleSpaceId).Result()
	if err != nil {
		player_mutex.Unlock()
		return
	}
	player_mutex.Unlock()
	if len(results) != 1 {
		err = errs.ErrorPlayerOnlineDataLost
		return
	}
	if results[0] != spaceid {
		err = errs.ErrorPlayerIsNotInBattleSpace
		return
	}

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err = space_mutex.Lock(); err != nil {
		return
	}
	defer space_mutex.Unlock()
	spaceResults, err := client.HGetAll(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {

		return
	}

	if !clientId.Equal(&network.ClientID{Address: spaceResults[rdsconst.PalyerMapClientIdAddress],
		Id: spaceResults[rdsconst.PlayerMapClientIdId]}) {
		err = errs.ErrorPermissionsLost
		return
	}

	if spaceResults[rdsconst.BattleSpaceState] == rdsconst.BattleSpaceStateRunning {
		err = errs.ErrorSpaceIsRunning
		return
	}
	pipe := client.TxPipeline()
	defer pipe.Close()

	// pipe.Do(ctx, "MULTI")

	pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpaceState, rdsconst.BattleSpaceStateRunning)

	// pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return
	}
	return
}

func GetBattleSpaceInfo(ctx context.Context, spaceid string)([]interface{},error){
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return nil,err
	}
	results, err := client.HMGet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid),
	 rdsconst.BattleSpaceMasterUid,
	 rdsconst.BattleSpacePlayerCount,
	).Result()

	if err != nil {
		return nil,err
	}
	if len(results) != 2 {
		return nil,errs.ErrorSpaceOnlineDataLost
	}
	if results[0] == nil {
		return nil,errs.ErrorSpaceOnlineDataLost
	}
	if results[1] == nil {
		return nil,errs.ErrorSpaceOnlineDataLost
	}
	return results,nil
}
func LeaveBattleSpace(ctx context.Context, clientId *network.ClientID) (uid string, spaceid string, err error) {
	uid, err = client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		return
	}

	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err = player_mutex.Lock(); err != nil {
		return
	}

	playerResults, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientBattleSpaceId).Result()
	if err != nil {
		player_mutex.Unlock()
		return
	}
	player_mutex.Unlock()
	if len(playerResults) != 1 {
		err = errs.ErrorPlayerOnlineDataLost
		return
	}
	if playerResults[0] == nil || playerResults[0].(string) == "" {
		err = errs.ErrorPlayerIsNotInBattleSpace
		return
	}

	spaceid = playerResults[0].(string)
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err = space_mutex.Lock(); err != nil {
		return
	}
	defer space_mutex.Unlock()
	results, err := client.HMGet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePlayerPos).Result()
	if err != nil {

		return
	}

	if len(results) != 1 {
		err = errs.ErrorSpaceOnlineDataLost
		return
	}
	if results[0] == nil {
		err = errs.ErrorSpaceOnlineDataLost
		return
	}

	player_pos := results[0].(string)
	player_pos = rdsconst.ReplaceData(rdsconst.SplitData(player_pos), uid, "")
	pipe := client.TxPipeline()
	defer pipe.Close()

	// pipe.Do(ctx, "MULTI")

	pipe.HDel(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.GetBattleSpacePlayerDataKey(uid))
	pipe.HMSet(ctx, rdsconst.GetPlayerOnlineDataKey(uid), rdsconst.PlayerMapClientBattleSpaceId, nil)
	pipe.HMSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePlayerPos, player_pos)
	// pipe.Do(ctx, "exec")

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

func BattleSpaceReportNat(ctx context.Context, spaceid string, nat string, clientId *network.ClientID) error {
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return err
	}
	defer space_mutex.Unlock()
	results, err := client.HMGet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpaceNatAddr, rdsconst.PalyerMapClientIdAddress, rdsconst.PlayerMapClientIdId).Result()
	if err != nil {

		return err
	}
	if len(results) != 3 {
		return errs.ErrorPlayerOnlineDataLost
	}
	if !clientId.Equal(&network.ClientID{Address: results[1].(string), Id: results[2].(string)}) {
		return errs.ErrorPermissionsLost
	}
	if results[0].(string) != "" {
		return errs.ErrorSpaceOnlineDataLost
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	// pipe.Do(ctx, "MULTI")

	pipe.HSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpaceNatAddr, nat)

	// pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}
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

	results, err := client.HMGet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpacePlayerPos).Result()
	if err != nil {
		space_mutex.Unlock()
		return nil
	}
	space_mutex.Unlock()
	if len(results) != 1 {
		return nil
	}
	if results[0] == nil || results[0].(string) == "" {
		return nil
	}

	list := make([]*network.ClientID, 0)
	for _, uid := range rdsconst.SplitData(results[0].(string)) {
		if uid == "" {
			continue
		}
		cli := getBattleSpacePlayerClientID(ctx, uid)
		if cli == nil {
			continue
		}
		list = append(list, cli)
	}
	return list
}

func FindBattleSpaceData(ctx context.Context, spaceid string) (map[string]string, error) {
	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return nil, err
	}
	defer space_mutex.Unlock()
	results, err := client.HGetAll(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid)).Result()
	if err != nil {
		return nil, err
	}

	return results, nil
}

func FindBattleSpaceIDByClientID(ctx context.Context, clientId *network.ClientID) (string, error) {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		return "", err
	}

	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := player_mutex.Lock(); err != nil {
		return "", err
	}
	results, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientBattleSpaceId).Result()
	if err != nil {
		player_mutex.Unlock()
		return "", err
	}
	player_mutex.Unlock()
	if len(results) != 1 {
		return "", errors.New("player online data error")
	}
	if results[0] == nil || results[0].(string) == "" {
		return "", errs.ErrorPlayerIsNotInBattleSpace
	}
	return results[0].(string), nil
}

func IsMaster(ctx context.Context, clientId *network.ClientID) (bool, error) {

	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := player_mutex.Lock(); err != nil {
		return false, err
	}

	playerResults, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientBattleSpaceId).Result()
	if err != nil {
		player_mutex.Unlock()
		return false, err
	}
	player_mutex.Unlock()

	if len(playerResults) != 1 {
		return false, errs.ErrorPlayerOnlineDataLost
	}
	if playerResults[0] == nil || playerResults[0].(string) == "" {
		return false, errs.ErrorPlayerIsNotInBattleSpace
	}

	spaceid := playerResults[0].(string)

	space_mutex := sync.NewMutex(rdsconst.GetBattleSpaceLockKey(spaceid))
	if err := space_mutex.Lock(); err != nil {
		return false, err
	}
	defer space_mutex.Unlock()
	results, err := client.HMGet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(spaceid), rdsconst.BattleSpaceMasterUid).Result()
	if err != nil {
		return false, err
	}

	if len(results) != 1 {
		return false, errs.ErrorPlayerOnlineDataLost
	}
	if results[0] == nil {
		return false, errs.ErrorPlayerOnlineDataLost
	}
	masterid := results[0].(string)

	return masterid == uid, nil
}

func getBattleSpacePlayerClientID(ctx context.Context, uid string) *network.ClientID {
	player_mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := player_mutex.Lock(); err != nil {
		return nil
	}
	results, err := client.HMGet(ctx,
		rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PalyerMapClientIdAddress,
		rdsconst.PlayerMapClientIdId).Result()
	if err != nil {
		player_mutex.Unlock()
		return nil
	}
	player_mutex.Unlock()

	if len(results) != 2 {
		return nil
	}

	if results[0] == nil || results[1] == nil {
		return nil
	}
	return &network.ClientID{Address: results[0].(string), Id: results[1].(string)}
}
