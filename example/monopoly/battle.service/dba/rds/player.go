package rds

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"

	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	"github.com/yamakiller/velcro-go/network"
)

// GetPlayerUID 获取客户端绑定的UID
func GetPlayerUID(ctx context.Context, clientId *network.ClientID) (string, error) {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}

		return "", err
	}

	mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return "", err
	}

	results, err := client.HGetAll(ctx, rdsconst.GetPlayerOnlineDataKey(uid)).Result()

	if err == redis.Nil {
		mutex.Unlock()
		return "", nil
	}

	// 拥有者权限检查
	if !clientId.Equals(&network.ClientID{Address: results[rdsconst.PalyerMapClientIdAddress],
		ID: results[rdsconst.PlayerMapClientIdId]}) {
		mutex.Unlock()
		return "", errors.New("permissions lost")
	}
	mutex.Unlock()
	
	return uid, nil
}