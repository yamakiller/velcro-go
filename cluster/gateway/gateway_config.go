package gateway

import (
	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/encryption/ecdh"
	"github.com/yamakiller/velcro-go/utils/files"
)

// GatewayConfigOption 是一个配置网关参数的函数
type GatewayConfigOption func(option *GatewayConfig)

func Configure(options ...GatewayConfigOption) *GatewayConfig {
	config := defaultGatewayConfig()
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

func WithLAddrServant(laddr string) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.LAddrServant = laddr
	}
}

func WithNewSystem(newFunc func(options ...network.ConfigOption) *network.NetworkSystem) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.NewNetworkSystem = newFunc
	}
}

// WithClientPool 设置客户端池
func WithClientPool(pool GatewayClientPool) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.ClientPool = pool
	}
}

// WithNewEncryption 设置密钥交换算法对象创建函数
func WithNewEncryption(newFunc func() *Encryption) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.NewEncryption = newFunc
	}
}

// WithNetworkTimeout 设置客户端网络连接保活时间
func WithKleepalive(timeout int32) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.Kleepalive = timeout
	}
}

// WithKleepaliveServant 设置内部服务网络连接保活时间
func WithKleepaliveServant(timeout int32) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.KleepaliveServant = timeout
	}
}

// WithRequestTimeout 请求默认超时
func WithRequestTimeout(timeout int64) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.RequestDefautTimeout = timeout
	}
}

// WithOnlineOfNumber 设置最大在线人数
func WithOnlineOfNumber(number int) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.OnlineOfNumber = number
	}
}

// WithRoute 设置路由配置
func WithRoute(router *router.RouterConfig) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.Router = router
	}
}

// GatewayConfig 网关配置信息
type GatewayConfig struct {
	VAddr                string
	LAddr                string
	LAddrServant         string
	NewNetworkSystem     func(options ...network.ConfigOption) *network.NetworkSystem
	ClientPool           GatewayClientPool
	NewEncryption        func() *Encryption
	Kleepalive           int32
	KleepaliveServant    int32
	RequestDefautTimeout int64
	OnlineOfNumber       int
	Router               *router.RouterConfig
}

func defaultGatewayConfig() *GatewayConfig {
	return &GatewayConfig{
		NewNetworkSystem:     network.NewTCPServerNetworkSystem,
		NewEncryption:        defaultEncryption,
		Kleepalive:           2000,
		KleepaliveServant:    2000,
		RequestDefautTimeout: 1000,
		OnlineOfNumber:       2000,
		Router: &router.RouterConfig{
			URI:       files.NewLocalPathFull("routes.yaml"),
			Algorithm: "p2c",
		},
	}
}

func defaultEncryption() *Encryption {
	return &Encryption{Ecdh: &ecdh.Curve25519{A: 247, B: 127, C: 64}}
}
