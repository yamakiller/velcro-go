package rds

const (
	player_lock        = "player_lock_"        // key
	player_uid         = "player_uid_"         // key
	player_client_id   = "player_client_id_"   // clientid to string()
	player_online_data = "player_online_data_" // key

	player_online_table = "player_online_table" // 用户在线表

	player_map_client_uid        = "uid"
	palyer_map_client_id_address = "client_id_address"
	player_map_client_id_id      = "client_id_id"
)

// getPlayerLockKey 获取player级共享锁Key
func getPlayerLockKey(key string) string {
	return player_lock + key
}

// getPlayerUidKey player uid => client id
func getPlayerUidKey(key string) string {
	return player_uid + key
}

// getplayerClientIDKey client id => key
func getPlayerClientIDKey(clientId string) string {
	return player_client_id + clientId
}

// getPlayerOnlineDataKey player 在线数据的key
func getPlayerOnlineDataKey(key string) string {
	return player_online_data + key
}
