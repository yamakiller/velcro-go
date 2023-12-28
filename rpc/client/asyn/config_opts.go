package asyn

import "google.golang.org/protobuf/proto"

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

func WithConnected(f func()) ConnConfigOption {
	return func(config *ConnConfig) {
		config.Connected = f
	}
}

func WithReceive(f func(proto.Message)) ConnConfigOption {
	return func(config *ConnConfig) {
		config.Receive = f
	}
}

func WithClosed(f func()) ConnConfigOption {
	return func(config *ConnConfig) {
		config.Closed = f
	}
}
