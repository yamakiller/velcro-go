package network

import (
	"github.com/yamakiller/velcro-go/logs"
	"go.opentelemetry.io/otel/metric"
)

// ConfigOption 是一个配置Network系统的函数
type ConfigOption func(config *Config)

func Configure(options ...ConfigOption) *Config {
	config := defaultConfig()
	for _, option := range options {
		option(config)
	}

	return config
}

// WithMetricProviders sets the metric providers
func WithMetricProviders(provider metric.MeterProvider) ConfigOption {

	return func(config *Config) {
		config.MetricsProvider = provider
	}
}

// WithDefaultPrometheusProvider sets the default prometheus provider
func WithDefaultPrometheusProvider(port ...int) ConfigOption {
	_port := 2222
	if len(port) > 0 {
		_port = port[0]
	}

	return WithMetricProviders(defaultPrometheusProvider(_port))
}

// WithLoggerFactory sets the logger factory to use for the actor system
func WithLoggerFactory(factory func(system *NetworkSystem) logs.LogAgent) ConfigOption {
	return func(config *Config) {
		config.LoggerFactory = factory
	}
}

func WithProducer(producer ProducerWidthClientSystem) ConfigOption {
	return func(config *Config) {
		config.Producer = producer
	}
}
