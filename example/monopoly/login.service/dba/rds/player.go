package rds

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/errs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/sign"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsstruct"
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
	defer mutex.Unlock()
	n, err := client.Exists(ctx, rdsconst.GetPlayerUidKey(uid)).Result()
	if err != nil {
		return err
	}

	oldClientID := ""
	if n > 0 {
		oldClientID, err = client.Get(ctx, rdsconst.GetPlayerUidKey(uid)).Result()
		if err != nil {
			return err
		}
	}

	oldClient := &rdsstruct.RdsPlayerData{}

	client.Get(ctx, rdsconst.GetPlayerOnlineDataKey(uid)).Scan(oldClient)
	// if oldClient == nil {
	// 	oldClient = &rdsstruct.RdsPlayerData{}
	// }
	if oldClient.Externs == nil {
		oldClient.Externs = make(map[string]string)
	}
	oldClient.DisplayName = displayName
	oldClient.UID = uid
	oldClient.ClientIdAddress = clientId.Address
	oldClient.ClientIdId = clientId.Id
	for k, v := range account.Externs {
		oldClient.Externs[k] = v
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	// pipe.Do(ctx, "MULTI")

	pipe.Set(ctx, rdsconst.GetPlayerUidKey(uid), clientId.ToString(), 0)
	pipe.Set(ctx, rdsconst.GetPlayerOnlineDataKey(uid), oldClient, 0)

	if oldClientID != clientId.ToString() && oldClientID != "" {
		pipe.Del(ctx, rdsconst.GetPlayerClientIDKey(oldClientID))
	} else {
		pipe.RPush(ctx, rdsconst.PlayerOnlineTable, uid)
	}

	pipe.Set(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString()), uid, 0)

	// pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}

	// 加入监视器，监视getPlayerUidKey(uid) 超时删除
	return nil
}

func UnRegisterPlayer(ctx context.Context, clientId *network.ClientID) error {

	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return err
	}
	defer mutex.Unlock()

	oldClient := &rdsstruct.RdsPlayerData{}
	if err := client.Get(ctx, rdsconst.GetPlayerOnlineDataKey(uid)).Scan(oldClient); err != nil {
		return err
	}

	// if oldClient == nil {
	// 	return errs.ErrUnRegisterPlayerNoExsit
	// }

	// 拥有者权限检查
	if !clientId.Equal(&network.ClientID{Address: oldClient.ClientIdAddress,
		Id: oldClient.ClientIdId}) {
		mutex.Unlock()
		return errs.ErrPermissionsLost
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	// pipe.Do(ctx, "MULTI")

	pipe.Del(ctx, rdsconst.GetPlayerUidKey(uid))
	pipe.Del(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString()))
	pipe.LRem(ctx, rdsconst.PlayerOnlineTable, 1, uid)

	pipe.Del(ctx, rdsconst.GetPlayerOnlineDataKey(uid))

	// pipe.Do(ctx, "exec")

	_, err = pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}

	return nil
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
	oldClient := &rdsstruct.RdsPlayerData{}
	if err := client.Get(ctx, rdsconst.GetPlayerOnlineDataKey(uid)).Scan(oldClient); err != nil {
		return false, err
	}

	// if oldClient == nil {
	// 	return false, errs.ErrUnRegisterPlayerNoExsit
	// }

	if !clientId.Equal(&network.ClientID{Address: oldClient.ClientIdAddress, Id: oldClient.ClientIdId}) {
		return false, nil
	}
	return true, nil
}

func FindPlayerData(ctx context.Context, clientId *network.ClientID) (*rdsstruct.RdsPlayerData, error) {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}

	mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}

	defer mutex.Unlock()
	oldClient := &rdsstruct.RdsPlayerData{}
	if err := client.Get(ctx, rdsconst.GetPlayerOnlineDataKey(uid)).Scan(oldClient); err != nil {
		return nil, err
	}

	// if oldClient == nil {
	// 	return nil, errs.ErrUnRegisterPlayerNoExsit
	// }
	return oldClient, nil
}

func GetPlayerBattleSpaceID(ctx context.Context, uid string) (string, error) {
	mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return "", err
	}

	defer mutex.Unlock()
	return client.Get(ctx, rdsconst.GetPlayerBattleSpaceIDKey(uid)).Result()
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
