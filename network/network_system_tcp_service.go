package network

import (
	"github.com/lithammer/shortuuid/v4"
	"github.com/yamakiller/velcro-go/extensions"
)

func NewTCPNetworkSystem(options ...ConfigOption) *NetworkSystem {
	config := Configure(options...)

	return NewTCPNetworkSystemConfig(config)
}

func NewTCPNetworkSystemConfig(config *Config) *NetworkSystem {
	ns := &NetworkSystem{NetType: "TCPServer"}
	ns.ID = shortuuid.New()
	ns._producer = config.Producer
	ns._handlers = NewHandlerRegistry(ns)
	ns._extensions = extensions.NewExtensions()
	ns._logger = config.LoggerFactory(ns)

	ns._module = newTcpNetworkServerModule(ns)

	return ns
}
