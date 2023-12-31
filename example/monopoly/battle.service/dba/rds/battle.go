package rds

import (
	"context"
	"strconv"
	"time"

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
	mapURi string) error {

	mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return err
	}

	results, err := client.HMGet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PlayerMapClientIcon,
		rdsconst.PlayerMapClientBattleSpaceId).Result()
	if err != nil {
		mutex.Unlock()
		return err
	}

	if len(results) != 2 {
		mutex.Unlock()
		return errs.ErrorPlayerOnlineDataLost
	}

	if results[1] != "" {
		mutex.Unlock()
		return errs.ErrorPlayerAlreadyInBattleSpace
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Do(ctx, "MULTI")

	results[1] = shortuuid.New()

	battleSpace := map[string]string{
		rdsconst.BattleSpaceId:            results[0].(string),
		rdsconst.PalyerMapClientIdAddress: clientId.Address,
		rdsconst.PlayerMapClientIdId:      clientId.Id,
		rdsconst.BattleSpaceMapURi:        mapURi,
		rdsconst.BattleSpaceNatAddr:       "",
		rdsconst.BattleSpacePos1Uid:       uid,
		rdsconst.BattleSpacePos2Uid:       "",
		rdsconst.BattleSpacePos3Uid:       "",
		rdsconst.BattleSpacePos4Uid:       "",
		rdsconst.BattleSpacePos1Icon:      results[0].(string),
		rdsconst.BattleSpacePos2Icon:      results[0].(string),
		rdsconst.BattleSpacePos3Icon:      results[0].(string),
		rdsconst.BattleSpacePos4Icon:      results[0].(string),
		rdsconst.BattleSpaceTime:          strconv.FormatInt(time.Now().UnixMilli(), 10),
		rdsconst.BateleSpaceState:         rdsconst.BattleSpaceStateNomal,
	}

	pipe.HSet(ctx, rdsconst.GetPlayerOnlineDataKey(uid), rdsconst.PlayerMapClientBattleSpaceId, results[1])
	pipe.HMSet(ctx, rdsconst.GetBattleSpaceOnlineDataKey(results[1].(string)), battleSpace)
	pipe.RPush(ctx, rdsconst.BattleSpaceOnlinetable, results[1].(string))

	pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		mutex.Unlock()
		return err
	}

	mutex.Unlock()

	return nil
}
