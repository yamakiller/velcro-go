package rds

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/errs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/sign"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	"github.com/yamakiller/velcro-go/network"
)

// 房间在Redis中的存储关系

func RegisterPlayer(ctx context.Context,
	clientId *network.ClientID,
	uid string,
	displayName string,
	account *sign.Account) error {

	mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return err
	}

	n, err := client.Exists(ctx, rdsconst.GetPlayerUidKey(uid)).Result()
	if err != nil {
		mutex.Unlock()
		return err
	}

	oldClientID := ""
	if n > 0 {
		oldClientID, err = client.Get(ctx, rdsconst.GetPlayerUidKey(uid)).Result()
		if err != nil {
			mutex.Unlock()
			return err
		}
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	// pipe.Do(ctx, "MULTI")

	pipe.Set(ctx, rdsconst.GetPlayerUidKey(uid), clientId.ToString(), 0)
	pipe.Set(ctx, rdsconst.GetPlayerDisplayNameKey(displayName), uid, 0)

	if oldClientID != clientId.ToString() && oldClientID != "" {
		pipe.Del(ctx, rdsconst.GetPlayerClientIDKey(oldClientID))
	} else {
		pipe.RPush(ctx, rdsconst.PlayerOnlineTable, uid)
	}

	pipe.Set(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString()), uid, 0)

	if oldClientID != clientId.ToString() && oldClientID != "" {
		// 如果用户已经存在更形client id 信息;
		pipe.HSet(ctx, rdsconst.GetPlayerOnlineDataKey(uid),
			rdsconst.PalyerMapClientIdAddress, clientId.Address,
			rdsconst.PlayerMapClientIdId, clientId.Id)
	} else {
		// 如果用户不存在,插入在线数据;
		member := map[string]string{
			rdsconst.PlayerMapClientUid:         uid,
			rdsconst.PlayerMapClientDisplayName: displayName,
			rdsconst.PalyerMapClientIdAddress:   clientId.Address,
			rdsconst.PlayerMapClientIdId:        clientId.Id,
		}
		for mkey, mvalue := range account.Externs {
			member[mkey] = mvalue
		}

		pipe.HMSet(ctx, rdsconst.GetPlayerOnlineDataKey(uid), member)
	}

	// pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		mutex.Unlock()
		return err
	}

	mutex.Unlock()

	// 加入监视器，监视getPlayerUidKey(uid) 超时删除
	return nil
}

func UnRegisterPlayer(ctx context.Context, clientId *network.ClientID, uid string) (map[string]string, error) {

	mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}

	results, err := client.HGetAll(ctx, rdsconst.GetPlayerOnlineDataKey(uid)).Result()

	if err == redis.Nil {
		mutex.Unlock()
		return nil, nil
	}

	displayName := results[rdsconst.PlayerMapClientDisplayName]

	// 拥有者权限检查
	if !clientId.Equal(&network.ClientID{Address: results[rdsconst.PalyerMapClientIdAddress],
		Id: results[rdsconst.PlayerMapClientIdId]}) {
		mutex.Unlock()
		return nil, errs.ErrPermissionsLost
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Do(ctx, "MULTI")

	pipe.Del(ctx, rdsconst.GetPlayerUidKey(uid))
	pipe.Del(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString()))
	pipe.Del(ctx, rdsconst.GetPlayerDisplayNameKey(displayName))
	pipe.LRem(ctx, rdsconst.PlayerOnlineTable, 1, uid)

	pipe.Del(ctx, rdsconst.GetPlayerOnlineDataKey(uid))

	pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		mutex.Unlock()
		return nil, err
	}

	mutex.Unlock()

	return results, nil
}

func IsAuth(ctx context.Context, clientId *network.ClientID) (bool, error) {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}

		return false, err
	}

	mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return false, err
	}

	values, err := client.HMGet(ctx,
		rdsconst.GetPlayerOnlineDataKey(uid),
		rdsconst.PalyerMapClientIdAddress,
		rdsconst.PlayerMapClientIdId).Result()
	if err != nil {
		if err == redis.Nil {
			mutex.Unlock()
			return false, nil
		}
		mutex.Unlock()
		return false, err
	}
	mutex.Unlock()

	if len(values) != 2 {
		return false, nil
	}

	if !clientId.Equal(&network.ClientID{Address: values[0].(string), Id: values[1].(string)}) {
		return false, nil
	}
	return true, nil
}

// GetPlayerUID 获取客户端绑定的UID
func GetPlayerUID(ctx context.Context, clientId *network.ClientID) (string, error) {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}

		return "", err
	}

	return uid, nil
}

/*func GetUserCount(ctx context.Context) (int64, error) {
	val, err := client.LLen(ctx, getUserArrayKey("")).Result()
	if err != nil {
		return 0, err
	}

	return val, nil
}*/

/*func GetUserRange(startIndex, endIndex int64) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), setTimeout)
	defer cancel()

	val, err := _client.LRange(ctx, getUserArrayKey(""), startIndex, endIndex).Result()
	if err != nil {
		return nil, err
	}

	return val, nil
}

func RemUser(uid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), setTimeout)
	defer cancel()

	pipe := _client.TxPipeline()
	defer pipe.Close()

	pipe.LRem(ctx, getUserArrayKey(uid), 0, uid)

	_, err := pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}

	return nil
}

func getUserArrayKey(uid string) string {
	return fmt.Sprintf("user_array:%s", uid)
}*/
