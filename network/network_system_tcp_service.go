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
	ns.producer = config.Producer
	ns.handlers = NewHandlerRegistry(ns, config.VAddr)
	ns.extensions = extensions.NewExtensions()
	ns.extensionId = extensions.NextExtensionID()
	ns.logger = config.LoggerFactory(ns)
	ns.module = newTCPNetworkServerModule(ns)

	return ns
}
