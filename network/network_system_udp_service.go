package network

import (
	"github.com/lithammer/shortuuid/v4"
	"github.com/yamakiller/velcro-go/extensions"
)

func NewUDPNetworkSystem(options ...ConfigOption) *NetworkSystem {
	config := Configure(options...)

	return NewTCPNetworkSystemConfig(config)
}

func NewUDPNetworkSystemConfig(config *Config) *NetworkSystem {
	ns := &NetworkSystem{}
	ns.ID = shortuuid.New()
	ns.Config = config
	if ns.Config.MetricsProvider != nil {
		ns.Config.meriicsKey = "udpserver" + ns.ID
	}
	ns._producer = config.Producer
	ns._handlers = NewHandlerRegistry(ns)
	ns._extensions = extensions.NewExtensions()
	ns._extensionId = extensions.NextExtensionID()
	ns._logger = config.LoggerFactory(ns)
	ns._module = newTcpNetworkServerModule(ns)

	return ns
}
