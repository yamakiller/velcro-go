package network

import (
	"github.com/lithammer/shortuuid/v4"
)

func NewWSServerNetworkSystem(options ...ConfigOption) *NetworkSystem {
	config := Configure(options...)

	return NewWSServerNetworkSystemConfig(config)
}

func NewWSServerNetworkSystemConfig(config *Config) *NetworkSystem {
	ns := &NetworkSystem{}
	ns.ID = shortuuid.New()
	ns.Config = config
	ns.producer = config.Producer
	ns.handlers = NewHandlerRegistry(ns, config.VAddr)
	ns.module = newWSNetworkServerModule(ns)

	return ns
}
