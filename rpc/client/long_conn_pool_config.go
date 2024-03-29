package client

import "time"

type LongConnPoolConfig struct {
	MaxConn            int32         // 最大连接数
	MaxIdleConn        int32         // 最大空闲连接数
	MaxIdleConnTimeout time.Duration // 空闲连接最大闲置时间
	ConnectTimeout     time.Duration // 连接超时时间
}

const (
	defaultMaxConn            = 1 << 20 // no limit
	defaultMaxIdleConn        = 1
	defaultMaxIdleConnTimeout = 30 * time.Minute
	defaultConnectTimeout     = 2 * time.Second
)

func CheckLongConnPoolConfig(config LongConnPoolConfig) *LongConnPoolConfig {
	if config.MaxConn <= 0 {
		config.MaxConn = defaultMaxConn
	}

	if config.MaxIdleConn <= 0 {
		config.MaxIdleConn = defaultMaxIdleConn
	}

	if config.MaxIdleConnTimeout <= 0 {
		config.MaxIdleConnTimeout = defaultMaxIdleConnTimeout
	}

	if config.ConnectTimeout <= 0 {
		config.ConnectTimeout = defaultConnectTimeout
	}
	return &config
}
