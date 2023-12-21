package service

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
}

func WithAddr(addr string)ServiceConfigOption{
	return func(option *ServiceConfig) {
		option.Addr = addr
	}
}

func defaultServiceConfig() * ServiceConfig{
	return &ServiceConfig{

	}
}