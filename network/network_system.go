package network

import (
	"github.com/yamakiller/velcro-go/extensions"
	"github.com/yamakiller/velcro-go/logs"
)

// NetworkSystem 网络系统
type NetworkSystem struct {
	Config  *Config
	ID      string
	NetType string

	_module     NetworkModule
	_producer   ProducerWidthClientSystem
	_handlers   *HandleRegistryValue
	_extensions *extensions.Extensions
	_logger     logs.LogAgent // 日志代理接口
}

func (ns *NetworkSystem) Address() string {
	return ns._handlers.Address
}

func (ns *NetworkSystem) Open(addr string) error {
	return ns._module.Open(addr)
}

func (ns *NetworkSystem) Shutdown() {
	ns._module.Stop()
}

func (ns *NetworkSystem) Logger() logs.LogAgent {
	return ns._logger
}

/*func (ns *NetworkSystem) IsStopped() bool {
	select {
	case <-ns._stopper:
		return true
	default:
		return false
	}
}*/
