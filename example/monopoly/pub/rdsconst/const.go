package rdsconst

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
	battle_space_online_data = "battle_space_online_data_"
	BattleSpaceId            = "space_id"
	BattleSpaceMapURi        = "space_map_uri"
	BattleSpaceTime          = "space_time"
	BattleSpaceNatAddr       = "space_nat_addr"
	BateleSpaceState         = "space_state"
	BattleSpacePos1Uid       = "space_pos_1_uid"
	BattleSpacePos2Uid       = "space_pos_2_uid"
	BattleSpacePos3Uid       = "space_pos_3_uid"
	BattleSpacePos4Uid       = "space_pos_4_uid"

	BattleSpacePos1Icon = "space_pos_1_icon"
	BattleSpacePos2Icon = "space_pos_2_icon"
	BattleSpacePos3Icon = "space_pos_3_icon"
	BattleSpacePos4Icon = "space_pos_4_icon"

	BattleSpaceStateNomal   = "nomal"   // 正常状态
	BattleSpaceStateNat     = "nat"     // 接收nat信息状态
	BattleSpaceStateReady   = "ready"   // 就绪状态
	BattleSpaceStateRunning = "running" // 游戏中/运行中
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
