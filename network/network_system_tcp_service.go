package network

import (
	"github.com/lithammer/shortuuid/v4"
	"github.com/yamakiller/velcro-go/extensions"
)

func NewTCPServerNetworkSystem(options ...ConfigOption) *NetworkSystem {
	config := Configure(options...)

	return NewTCPServerNetworkSystemConfig(config)
}

func NewTCPServerNetworkSystemConfig(config *Config) *NetworkSystem {
	ns := &NetworkSystem{}
	ns.ID = shortuuid.New()
	ns.Config = config
	if ns.Config.MetricsProvider != nil {
		ns.Config.meriicsKey = "tcpserver" + ns.ID
	}
	ns._producer = config.Producer
	ns._handlers = NewHandlerRegistry(ns)
	ns._extensions = extensions.NewExtensions()
	ns._extensionId = extensions.NextExtensionID()
	ns._logger = config.LoggerFactory(ns)
	ns._module = newTCPNetworkServerModule(ns)

	return ns
}
