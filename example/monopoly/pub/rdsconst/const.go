package rdsconst

import (
	"fmt"
	"strings"
)

const (
	player_lock        = "player_lock_"         // uid
	player_uid         = "player_uid_"          // uid
	player_dispay_name = "player_dispaly_name_" // display name to uid
	player_client_id   = "player_client_id_"    // string from clientid to uid
	player_online_data = "player_online_data_"  // uid

	PlayerOnlineTable      = "player_online_table"       // 用户在线表
	BattleSpaceOnlinetable = "battle_sapce_online_table" // 对战战空间在线表

	PlayerMapClientUid           = "uid"
	PlayerMapClientDisplayName   = "display_name"
	PalyerMapClientIdAddress     = "client_id_address"
	PlayerMapClientIdId          = "client_id_id"
	PlayerMapClientIcon          = "icon"
	PlayerMapClientBattleSpaceId = "battle_space_id"
)

const (
	battle_space_lock        = "battle_space_lock_"
	battle_space_online_data = "battle_space_online_data_"
	BattleSpaceId            = "space_id"
	BattleSpaceName          = "space_name"
	BattleSpaceMapURi        = "space_map_uri"
	BattleSpaceTime          = "space_time"
	BattleSpaceNatAddr       = "space_nat_addr"
	BattleSpaceState         = "space_state"
	BattleSpaceMasterUid     = "space_master_uid"
	BattleSpaceMasterIcon    = "space_master_icon"
	BattleSpaceMasterDisplay = "space_master_display"
	BattleSpacePlayerCount   = "space_player_count"
	BattleSpacePassword      = "space_player_password"
	BattleSpaceExtend        = "space_player_extend"
	BattleSpacePlayerPos     = "space_player_pos" //玩家位置
	BattleSpacePlayer        = "space_player_"    //玩家信息
	BattleSpaceStateNomal    = "nomal"            // 正常状态
	BattleSpaceStateNat      = "nat"              // 接收nat信息状态
	BattleSpaceStateReady    = "ready"            // 就绪状态
	BattleSpaceStateRunning  = "running"          // 游戏中/运行中
)

// GetPlayerLockKey 获取player级共享锁Key
func GetPlayerLockKey(uid string) string {
	return player_lock + uid
}

// GetPlayerUidKey player uid => client id
func GetPlayerUidKey(uid string) string {
	return player_uid + uid
}

// GetPlayerDisplayNameKey player display name => uid
func GetPlayerDisplayNameKey(displayName string) string {
	return player_dispay_name + displayName
}

// GetPlayerClientIDKey client id => key
func GetPlayerClientIDKey(clientId string) string {
	return player_client_id + clientId
}

// GetPlayerOnlineDataKey player 在线数据的key
func GetPlayerOnlineDataKey(uid string) string {
	return player_online_data + uid
}

// GetBattleSpaceOnlineDataKey
func GetBattleSpaceOnlineDataKey(spaceid string) string {
	return battle_space_online_data + spaceid
}

// GetBattleSpacePlayerDataKey
func GetBattleSpacePlayerDataKey(uid string) string {
	return BattleSpacePlayer + uid
}

// GetBattleSpaceLockKey 获取battle级共享锁Key
func GetBattleSpaceLockKey(id string) string {
	return battle_space_lock + id
}

func MakeData(list []string) string {
	res := ""
	for i := (0); i < len(list); i++ {
		res += fmt.Sprintf("%s&", list[i])
	}
	return res[:len(res)-1]
}

func UpdateData(list []string, index int, val string) string {
	for i := 0; i < len(list); i++ {
		if i == index {
			list[i] = val
		}
	}
	return MakeData(list)
}

func ReplaceData(list []string, old, new string) string {
	for i := 0; i < len(list); i++ {
		if list[i] == old {
			list[i] = new
		}
	}
	return MakeData(list)
}
func EnterData(list []string, val string) (string, int32) {
	index := int32(-1)
	for i := 0; i < len(list); i++ {
		if list[i] == "" {
			list[i] = val
			index = int32(i)
			break
		}
	}
	return MakeData(list), index
}
func SplitData(data string) []string {
	return strings.Split(data, "&")
}
