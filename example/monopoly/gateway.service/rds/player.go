package rds

import (
	"context"

	"github.com/yamakiller/velcro-go/example/monopoly/pub/rdsconst"
	"github.com/yamakiller/velcro-go/network"
)

func GetPlayerClientID(ctx context.Context, uid string) *network.ClientID {
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