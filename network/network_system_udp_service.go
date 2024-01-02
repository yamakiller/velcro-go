package network

import (
	"github.com/lithammer/shortuuid/v4"
)

func NewUDPNetworkSystem(options ...ConfigOption) *NetworkSystem {
	config := Configure(options...)

	return NewUDPNetworkSystemConfig(config)
}

func NewUDPNetworkSystemConfig(config *Config) *NetworkSystem {
	ns := &NetworkSystem{}
	ns.ID = shortuuid.New()
	ns.Config = config
	ns.producer = config.Producer
	ns.handlers = NewHandlerRegistry(ns, config.VAddr)
	ns.module = newUDPNetworkServerModule(ns)

	return ns
}
