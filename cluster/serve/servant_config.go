package serve

import (
	"github.com/yamakiller/velcro-go/cluster/router"
	"go.opentelemetry.io/otel/metric"
)

func defaultConfig() *ServantConfig {
	return &ServantConfig{
		Kleepalive: 10 * 1000,
	}
}

type ServantConfig struct {
	MetricsProvider metric.MeterProvider
	Producer        func(*ServantClientConn) ServantClientActor
	Name            string
	LAddr           string
	VAddr           string
	Kleepalive      int32
	Router          *router.RouterConfig
}

type ServantConfigOption func(config *ServantConfig)

func configure(options ...ServantConfigOption) *ServantConfig {
	config := defaultConfig()
	for _, option := range options {
		option(config)
	}

	return config
}

// WithKleepalive 设置心跳时间
func WithKleepalive(kleepalive int32) ServantConfigOption {

	return func(config *ServantConfig) {
		config.Kleepalive = kleepalive
	}
}

// WithMetricsProvider ...
func WithMetricsProvider(provider metric.MeterProvider) ServantConfigOption {
	return func(config *ServantConfig) {
		config.MetricsProvider = provider
	}
}

// WithProducerActor actor 创建器
func WithProducerActor(f func(*ServantClientConn) ServantClientActor) ServantConfigOption {
	return func(config *ServantConfig) {
		config.Producer = f
	}
}

// WithName 设置服务名称
func WithName(name string) ServantConfigOption {
	return func(config *ServantConfig) {
		config.Name = name
	}
}

// WithLAddr 设置本服务的监听地址
func WithLAddr(laddr string) ServantConfigOption {
	return func(config *ServantConfig) {
		config.LAddr = laddr
	}
}

// WithVAddr 设置本服务的虚地址
func WithVAddr(vaddr string) ServantConfigOption {
	return func(config *ServantConfig) {
		config.VAddr = vaddr
	}
}

// WithRoute 设置路由配置
func WithRoute(router *router.RouterConfig) ServantConfigOption {
	return func(opt *ServantConfig) {
		opt.Router = router
	}
}
