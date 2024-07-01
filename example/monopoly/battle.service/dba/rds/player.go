package rds

import (
	"context"

	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsstruct"
	"github.com/yamakiller/velcro-go/network"
)

// GetPlayerData 获取客户端绑定的PlayerData
func GetPlayerData(ctx context.Context, clientId *network.ClientID) (*rdsstruct.RdsPlayerData, error) {
	uid, err := client.Get(ctx, rdsconst.GetPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
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

	return oldClient, nil
}

// GetPlayerData 获取客户端绑定的PlayerData
func GetPlayerDataByUID(ctx context.Context, uid string) (*rdsstruct.RdsPlayerData, error) {
	mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}

	defer mutex.Unlock()

	oldClient := &rdsstruct.RdsPlayerData{}
	if err := client.Get(ctx, rdsconst.GetPlayerOnlineDataKey(uid)).Scan(oldClient); err != nil {
		return nil, err
	}

	return oldClient, nil
}

func GetBattleSpacePlayerClientID(ctx context.Context, uid string) *network.ClientID {
	player_data, err := GetPlayerDataByUID(ctx, uid)
	if err != nil {
		return nil
	}
	return &network.ClientID{Address: player_data.ClientIdAddress, Id: player_data.ClientIdId}
}

func GetPlayerBattleSpaceID(ctx context.Context, uid string) (string, error) {
	mutex := sync.NewMutex(rdsconst.GetPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return "", err
	}

	defer mutex.Unlock()
	return client.Get(ctx, rdsconst.GetPlayerBattleSpaceIDKey(uid)).Result()
}
