package rpcsynclient

import "github.com/yamakiller/velcro-go/rpc/rpcmessage"

// ConnConfigOption 是一个配置rpc connector 的函数
type ConnConfigOption func(config *ConnConfig)

func Configure(options ...ConnConfigOption) *ConnConfig {
	config := defaultConfig()
	for _, option := range options {
		option(config)
	}

	return config
}

func WithMarshalRequest(f rpcmessage.MarshalRequestFunc) ConnConfigOption {
	return func(config *ConnConfig) {
		config.MarshalRequest = f
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
