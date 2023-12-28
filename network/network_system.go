package network

import (
	"github.com/yamakiller/velcro-go/extensions"
	"github.com/yamakiller/velcro-go/logs"
)

// NetworkSystem 网络系统
type NetworkSystem struct {
	Config      *Config
	ID          string
	module      NetworkModule
	producer    ProducerWidthClientSystem
	handlers    *HandleRegistryValue
	extensions  *extensions.Extensions
	extensionId extensions.ExtensionID
	logger      logs.LogAgent // 日志代理接口
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

func (ns *NetworkSystem) getLogger() logs.LogAgent {
	return ns.logger
}

// 日志
func (ns *NetworkSystem) Info(sfmt string, args ...interface{}) {
	ns.logger.Info("["+ns.Config.VAddr+"/"+ns.module.Network()+"]("+ns.ID+")", sfmt, args...)
}

func (ns *NetworkSystem) Debug(sfmt string, args ...interface{}) {
	ns.logger.Debug("["+ns.Config.VAddr+"/"+ns.module.Network()+"]("+ns.ID+")", sfmt, args...)
}

func (ns *NetworkSystem) Error(sfmt string, args ...interface{}) {
	ns.logger.Error("["+ns.Config.VAddr+"/"+ns.module.Network()+"]("+ns.ID+")", sfmt, args...)
}

func (ns *NetworkSystem) Warning(sfmt string, args ...interface{}) {
	ns.logger.Warning("["+ns.Config.VAddr+"/"+ns.module.Network()+"]("+ns.ID+")", sfmt, args...)
}

func (ns *NetworkSystem) Fatal(sfmt string, args ...interface{}) {
	ns.logger.Fatal("["+ns.Config.VAddr+"/"+ns.module.Network()+"]("+ns.ID+")", sfmt, args...)
}

func (ns *NetworkSystem) Panic(sfmt string, args ...interface{}) {
	ns.logger.Panic("["+ns.Config.VAddr+"/"+ns.module.Network()+"]("+ns.ID+")", sfmt, args...)
}
