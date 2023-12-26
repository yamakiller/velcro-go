package rds

import (
	"context"
	"fmt"

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

func PushUser(uid string,player interface{}) error {

	ctx, cancel := context.WithTimeout(context.Background(), setTimeout)
	defer cancel()

	pipe := _client.TxPipeline()
	defer pipe.Close()

	pipe.RPush(ctx, getUserArrayKey(uid), player)

	_, err := pipe.Exec(ctx)
	if err != nil {
		pipe.Discard()
		return err
	}

	return nil
}

func GetUserCount() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), setTimeout)
	defer cancel()

	val, err := _client.LLen(ctx, getUserArrayKey("")).Result()
	if err != nil {
		return 0, err
	}

	return val, nil
}

func GetUserRange(startIndex, endIndex int64) ([]string, error) {
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
	return fmt.Sprintf("user_array:%s",uid) 
}
