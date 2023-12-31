package rds

type BattleSpace struct {
	ID      string   // 空间ID
	MapURI  string   // 地图地址
	NatAddr string   // Nat地址
	State   int      // 状态
	Player  []string // 参与者
}

type BattleSpacePlayer struct {
	UID         string
	DisplayName string
}
