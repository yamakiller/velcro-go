package network

import (
	"github.com/lithammer/shortuuid/v4"
	"github.com/yamakiller/velcro-go/extensions"
)

func NewTCPSyncServerNetworkSystem(options ...ConfigOption) *NetworkSystem {
	config := Configure(options...)

	return NewTCPServerNetworkSystemConfig(config)
}

func NewTCPSyncServerNetworkSystemConfig(config *Config) *NetworkSystem {
	ns := &NetworkSystem{}
	ns.ID = shortuuid.New()
	ns.Config = config
	if ns.Config.MetricsProvider != nil {
		ns.Config.meriicsKey = "tcp-sync-server" + ns.ID
	}
	ns.producer = config.Producer
	ns.handlers = NewHandlerRegistry(ns, config.VAddr)
	ns.extensions = extensions.NewExtensions()
	ns.extensionId = extensions.NextExtensionID()
	ns.logger = config.LoggerFactory(ns)
	ns.module = newTCPNetworkServerModule(ns)

	return ns
}
