package client

type LongConnPoolConfig struct {
	MaxConn            int32 // 最大连接数
	MaxIdleConn        int32 // 最大空闲连接数
	MaxIdleConnTimeout int32 // 空闲连接最大闲置时间
}
