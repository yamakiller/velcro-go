package network

import (
	"github.com/lithammer/shortuuid/v4"
	"github.com/yamakiller/velcro-go/extensions"
)

func NewTCPConnectorNetworkSystemConfig(config *Config) *NetworkSystem {
	ns := &NetworkSystem{NetType: "TCPServer"}
	ns.ID = shortuuid.New()
	ns.Config = config
	ns._producer = config.Producer
	ns._handlers = NewHandlerRegistry(ns)
	ns._extensions = extensions.NewExtensions()
	ns._logger = config.LoggerFactory(ns)

	ns._module = newTcpConnectorNetworkServerModule(ns)

	return ns
}
