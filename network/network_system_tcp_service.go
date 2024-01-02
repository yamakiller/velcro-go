package network

import (
	"github.com/lithammer/shortuuid/v4"
)

func NewTCPServerNetworkSystem(options ...ConfigOption) *NetworkSystem {
	config := Configure(options...)

	return NewTCPServerNetworkSystemConfig(config)
}

func NewTCPServerNetworkSystemConfig(config *Config) *NetworkSystem {
	ns := &NetworkSystem{}
	ns.ID = shortuuid.New()
	ns.Config = config
	ns.producer = config.Producer
	ns.handlers = NewHandlerRegistry(ns, config.VAddr)
	ns.module = newTCPNetworkServerModule(ns)

	return ns
}
