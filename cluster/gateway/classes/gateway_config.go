package classes

import "github.com/yamakiller/velcro-go/network"

// GatewayConfigOption 是一个配置网关参数的函数
type GatewayConfigOption func(option *GatewayConfig)

func Configure(options ...GatewayConfigOption) *GatewayConfig {
	config := defaultRouterRpcProxyOption()
	for _, option := range options {
		option(config)
	}

	return config
}

func WithVAddr(vaddr string) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.VAddr = vaddr
	}
}

func WithLAddr(laddr string) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.LAddr = laddr
	}
}

func WithSpawnSystem(spawn func() *network.NetworkSystem) GatewayConfigOption {

	return func(opt *GatewayConfig) {
		opt.NewNetworkSystem = spawn
	}
}

func WithSpawnEncryption(spawn func() *Encryption) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.NewEncryption = spawn
	}
}

func WithRouterURI(uri string) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.RouterURI = uri
	}
}

// GatewayConfig 网关配置信息
type GatewayConfig struct {
	VAddr            string
	LAddr            string
	NewNetworkSystem func() *network.NetworkSystem
	NewGroup         func() ClientGroup
	NewEncryption    func() *Encryption
	RouterURI        string
}

func defaultGatewayConfig() *GatewayConfig {
	return &GatewayConfig{
		NewNetworkSystem:
	}
}
