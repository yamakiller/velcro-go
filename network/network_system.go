package network

// NetworkSystem 网络系统
type NetworkSystem struct {
	Config   *Config
	ID       string
	module   NetworkModule
	producer ProducerWidthClientSystem
	handlers *HandleRegistryValue
}

func (ns *NetworkSystem) NewClientID(id string) *ClientID {
	return &ClientID{Address: ns.handlers.Address, Id: id}
}

func (ns *NetworkSystem) Address() string {
	return ns.handlers.Address
}

func (ns *NetworkSystem) Open(addr string) error {
	return ns.module.Open(addr)
}

func (ns *NetworkSystem) Shutdown() {
	ns.module.Stop()
}
