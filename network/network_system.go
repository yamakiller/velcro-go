package network

import (
	"github.com/yamakiller/velcro-go/extensions"
	"github.com/yamakiller/velcro-go/logs"
)

// NetworkSystem 网络系统
type NetworkSystem struct {
	Config       *Config
	ID           string
	_module      NetworkModule
	_producer    ProducerWidthClientSystem
	_handlers    *HandleRegistryValue
	_extensions  *extensions.Extensions
	_extensionId extensions.ExtensionID
	_logger      logs.LogAgent // 日志代理接口
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

func (ns *NetworkSystem) logger() logs.LogAgent {
	return ns._logger
}

// 日志
func (ns *NetworkSystem) Info(sfmt string, args ...interface{}) {
	ns._logger.Info("["+ns.Config.VAddr+"/"+ns._module.Network()+"]("+ns.ID+")", sfmt, args...)
}

func (ns *NetworkSystem) Debug(sfmt string, args ...interface{}) {
	ns._logger.Debug("["+ns.Config.VAddr+"/"+ns._module.Network()+"]("+ns.ID+")", sfmt, args...)
}

func (ns *NetworkSystem) Error(sfmt string, args ...interface{}) {
	ns._logger.Error("["+ns.Config.VAddr+"/"+ns._module.Network()+"]("+ns.ID+")", sfmt, args...)
}

func (ns *NetworkSystem) Warning(sfmt string, args ...interface{}) {
	ns._logger.Warning("["+ns.Config.VAddr+"/"+ns._module.Network()+"]("+ns.ID+")", sfmt, args...)
}

func (ns *NetworkSystem) Fatal(sfmt string, args ...interface{}) {
	ns._logger.Fatal("["+ns.Config.VAddr+"/"+ns._module.Network()+"]("+ns.ID+")", sfmt, args...)
}

func (ns *NetworkSystem) Panic(sfmt string, args ...interface{}) {
	ns._logger.Panic("["+ns.Config.VAddr+"/"+ns._module.Network()+"]("+ns.ID+")", sfmt, args...)
}
