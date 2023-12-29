package serve

import (
	"github.com/yamakiller/velcro-go/logs"
	"go.opentelemetry.io/otel/metric"
)

func defaultConfig() *ServantConfig {
	return &ServantConfig{
		Kleepalive: 10 * 1000,
	}
}

type ServantConfig struct {
	MetricsProvider metric.MeterProvider
	LoggerAgent     logs.LogAgent
	Producer        func(*ServantClientConn) ServantClientActor
	Kleepalive      int32
	VAddr           string
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

// WithVAddr 设置本服务的虚地址
func WithVAddr(vaddr string) ServantConfigOption {
	return func(config *ServantConfig) {
		config.VAddr = vaddr
	}
}

// WithLogger 设置关联日志文件
func WithLogger(agent logs.LogAgent) ServantConfigOption {
	return func(config *ServantConfig) {
		config.LoggerAgent = agent
	}
}
