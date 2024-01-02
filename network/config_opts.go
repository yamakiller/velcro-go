package network

// ConfigOption 是一个配置Network系统的函数
type ConfigOption func(config *Config)

func Configure(options ...ConfigOption) *Config {
	config := defaultConfig()
	for _, option := range options {
		option(config)
	}

	return config
}

func WithProducer(producer ProducerWidthClientSystem) ConfigOption {
	return func(config *Config) {
		config.Producer = producer
	}
}

// WithKleepalive 保活时间
func WithKleepalive(timeout int32) ConfigOption {
	return func(config *Config) {
		config.Kleepalive = timeout
	}
}

func WithVAddr(vaddr string) ConfigOption {
	return func(config *Config) {
		config.VAddr = vaddr
	}
}
