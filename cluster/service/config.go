package service

import "github.com/yamakiller/velcro-go/rpc/server"

type ServiceConfigOption func(option *ServiceConfig)

func Configure(options ...ServiceConfigOption) *ServiceConfig {
	config := defaultServiceConfig()
	for _, option := range options {
		option(config)
	}

	return config
}

type ServiceConfig struct {
	Addr string
	Pool server.RpcPool
}

func WithAddr(addr string) ServiceConfigOption {
	return func(option *ServiceConfig) {
		option.Addr = addr
	}
}

func WithPool(pool server.RpcPool) ServiceConfigOption {
	return func(option *ServiceConfig) {
		option.Pool = pool}
}

func defaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{}
}