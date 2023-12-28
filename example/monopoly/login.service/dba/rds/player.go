package rds

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts/sign"
	"github.com/yamakiller/velcro-go/network"
)

// 房间在Redis中的存储关系

/*func MakeRoom(room datas.Room) error {
	idRoomKey := getRoomKey(room.Id)

	ctx, cancel := context.WithTimeout(context.Background(), setTimeout)
	defer cancel()

	pipe := _client.TxPipeline()
	defer pipe.Close()

	pipe.RPush(ctx, getRoomListKey(), room.Id)

	roomess := map[string]interface{}{
		"RoomId":           room.Id,
		"RoomMap":          room.Map,
		"RoomOwner":        room.Owner,
		"RoomOwnerNatAddr": room.OwnerNatAddr,
		"RoomMembers":      room.Members,
		"RoomData":         room.Data,
		"RoomCreate":       room.Create,
	}
	pipe.HSet(ctx, idRoomKey, roomess)

	pipe.HSet(ctx, getPlayerIdKey(room.Members[0].Id), "RoomAddr", room.Id)

	_, err := pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}

	return nil
}*/

func RegisterPlayer(ctx context.Context,
	uid string,
	clientId *network.ClientID,
	account *sign.Account,
	onlineTimeout time.Duration) error {

	mutex := sync.NewMutex(getPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return err
	}

	n, err := client.Exists(ctx, getPlayerUidKey(uid)).Result()
	if err != nil {
		mutex.Unlock()
		return err
	}

	oldClientID := ""
	if n > 0 {
		oldClientID, err = client.Get(ctx, getPlayerUidKey(uid)).Result()
		if err != nil {
			mutex.Unlock()
			return err
		}
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Do(ctx, "MULTI")

	pipe.Set(ctx, getPlayerUidKey(uid), clientId.ToString(), 0)

	if oldClientID != clientId.ToString() && oldClientID != "" {
		pipe.Del(ctx, getPlayerClientIDKey(oldClientID))
	} else {
		pipe.RPush(ctx, player_online_table, uid)
	}

	pipe.Set(ctx, getPlayerClientIDKey(clientId.ToString()), uid, 0)

	if oldClientID != clientId.ToString() && oldClientID != "" {
		// 如果用户已经存在更形client id 信息;
		pipe.HSet(ctx, getPlayerOnlineDataKey(uid),
			palyer_map_client_id_address, clientId.Address,
			player_map_client_id_id, clientId.Id)
	} else {
		// 如果用户不存在,插入在线数据;
		member := map[string]string{
			player_map_client_uid:        uid,
			palyer_map_client_id_address: clientId.Address,
			player_map_client_id_id:      clientId.Id,
		}
		for mkey, mvalue := range account.Externs {
			member[mkey] = mvalue
		}

		pipe.HMSet(ctx, getPlayerOnlineDataKey(uid), member)
	}

	pipe.Expire(ctx, getPlayerUidKey(uid), onlineTimeout)

	pipe.Do(ctx, "exec")

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

func UnRegisterPlayer(ctx context.Context, clientId *network.ClientID) error {
	uid, err := client.Get(ctx, getPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}

		return err
	}

	mutex := sync.NewMutex(getPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return err
	}

	n, err := client.Exists(ctx, getPlayerUidKey(uid)).Result()
	if err != nil {
		mutex.Unlock()
		return err
	}

	oldClientID := ""
	if n > 0 {
		oldClientID, err = client.Get(ctx, getPlayerUidKey(uid)).Result()
		if err != nil {
			mutex.Unlock()
			return err
		}
	}

	pipe := client.TxPipeline()
	defer pipe.Close()

	pipe.Do(ctx, "MULTI")
	if oldClientID != clientId.ToString() && oldClientID != "" {
		pipe.Del(ctx, getPlayerClientIDKey(oldClientID))
	}else{
		pipe.Del(ctx, getPlayerClientIDKey(clientId.ToString()))
	}
	
	pipe.Del(ctx, getPlayerOnlineDataKey(uid))
	pipe.Del(ctx, getPlayerUidKey(uid))

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

func IsAuth(ctx context.Context, clientId *network.ClientID) (bool, error) {
	uid, err := client.Get(ctx, getPlayerClientIDKey(clientId.ToString())).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}

		return false, err
	}

	mutex := sync.NewMutex(getPlayerLockKey(uid))
	if err := mutex.Lock(); err != nil {
		return false, err
	}

	values, err := client.HMGet(ctx, getPlayerOnlineDataKey(uid), palyer_map_client_id_address, player_map_client_id_id).Result()
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
