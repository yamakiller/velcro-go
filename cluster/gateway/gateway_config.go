package gateway

import (
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/encryption/ecdh"
	"go.opentelemetry.io/otel/metric"
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

// WithRouterURI 设置路由配置文件地址
func WithRouterURI(uri string) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.RouterURI = uri
	}
}

// WithLoggerFactory 设置日志仓库函数
func WithLoggerFactory(factory func(system *network.NetworkSystem) logs.LogAgent) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.LoggerFactory = factory
	}
}

// WithNetworkTimeout 设置网络超时时间
func WithNetworkTimeout(timeout int32) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.NetowkTimeout = timeout
	}
}

// WithRouteProxyFrequency 设置路由代理检测频率
func WithRouteProxyFrequency(frequency int32) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.RouteProxyFrequency = frequency
	}
}

// WithRouteProxyFrequency 设置路由代理连接超时时间
func WithRouteProxyDialTimeout(dialTimeout int32) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.RouteProxyDialTimeout = dialTimeout
	}
}

// WithRouteProxyKleepalive 设置路由代理保活时间
func WithRouteProxyKleepalive(kleepalive int32) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.RouteProxyKleepalive = kleepalive
	}
}

// WithRouteProxyAlgorithm 设置路由代理平衡器算法s
func WithRouteProxyAlgorithm(algorithm string) GatewayConfigOption {
	return func(opt *GatewayConfig) {
		opt.RouteProxyAlgorithm = algorithm
	}
}

// GatewayConfig 网关配置信息
type GatewayConfig struct {
	VAddr            string
	LAddr            string
	NewNetworkSystem func(options ...network.ConfigOption) *network.NetworkSystem
	ClientPool       GatewayClientPool
	NewEncryption    func() *Encryption
	RouterURI        string
	MetricsProvider  metric.MeterProvider
	LoggerFactory    func(system *network.NetworkSystem) logs.LogAgent
	NetowkTimeout    int32

	RouteProxyFrequency   int32
	RouteProxyDialTimeout int32
	RouteProxyKleepalive  int32
	RouteProxyAlgorithm   string
}

func defaultGatewayConfig() *GatewayConfig {
	return &GatewayConfig{
		NewNetworkSystem: network.NewTCPServerNetworkSystem,
		NewEncryption:    defaultEncryption,
		MetricsProvider:  nil,
		LoggerFactory: func(system *network.NetworkSystem) logs.LogAgent {
			return ProduceLogger()
		},
		NetowkTimeout:         2000,
		RouteProxyFrequency:   2000,
		RouteProxyDialTimeout: 2000,
		RouteProxyKleepalive:  4000,
		RouteProxyAlgorithm:   "p2c",
	}
}

func defaultEncryption() *Encryption {
	return &Encryption{Ecdh: &ecdh.Curve25519{A: 247, B: 127, C: 64}}
}
