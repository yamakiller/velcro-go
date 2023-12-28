package serve

import "go.opentelemetry.io/otel/metric"

func defaultConfig() *ConnConfig {
	return &ConnConfig{
		Kleepalive: 10 * 1000,
	}
}

type ConnConfig struct {
	MetricsProvider metric.MeterProvider
	Producer        func(*ServantClientConn) ServantClientActor
	Kleepalive      int32
}

type ConnConfigOption func(config *ConnConfig)

func configure(options ...ConnConfigOption) *ConnConfig {
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

// WithMetricsProvider ...
func WithMetricsProvider(provider metric.MeterProvider) ConnConfigOption {
	return func(config *ConnConfig) {
		config.MetricsProvider = provider
	}
}

// WithProducerActor actor 创建器
func WithProducerActor(f func(*ServantClientConn) ServantClientActor) ConnConfigOption {
	return func(config *ConnConfig) {
		config.Producer = f
	}
}
