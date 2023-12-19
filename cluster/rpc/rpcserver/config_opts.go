package rpcserver

import (
	"github.com/yamakiller/velcro-go/cluster/rpc/rpcmessage"
)

// ConnConfigOption 是一个配置rpc connector 的函数
type ConnConfigOption func(config *ConnConfig)

func Configure(options ...ConnConfigOption) *ConnConfig {
	config := defaultConfig()
	for _, option := range options {
		option(config)
	}

	return config
}

// WithKleepalive 设置心跳时间
func WithKleepalive(kleepalive int32) ConnConfigOption {

	return func(config *ConnConfig) {
		config.Kleepalive = kleepalive
	}
}

func WithMarshalResponse(f rpcmessage.MarshalResponseFunc) ConnConfigOption {
	return func(config *ConnConfig) {
		config.MarshalResponse = f
	}
}

func WithMarshalMessage(f rpcmessage.MarshalMessageFunc) ConnConfigOption {
	return func(config *ConnConfig) {
		config.MarshalMessage = f
	}
}

func WithMarshalPing(f rpcmessage.MarshalPingFunc) ConnConfigOption {
	return func(config *ConnConfig) {
		config.MarshalPing = f
	}
}

func WithUnMarshal(f rpcmessage.UnMarshalFunc) ConnConfigOption {
	return func(config *ConnConfig) {
		config.UnMarshal = f
	}
}

func WithPool(f RpcPool) ConnConfigOption{
	return func(config *ConnConfig) {
		config.Pool = f
	}
}


