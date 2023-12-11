package network

import (
	"github.com/yamakiller/velcro-go/extensions"
	"github.com/yamakiller/velcro-go/logs"
)

func NewNetworkSystem() *NetworkSystem {
	ns := &NetworkSystem{_stopper: make(chan struct{})}
	/*switch typ {
	case TCPS:
		ns._module = newTcpNetworkServer(ns)
		break
	default:
		panic(errors.New("unknown system type"))
	}*/

	return ns
}

// NetworkSystem 网络系统
type NetworkSystem struct {
	Module      NetworkModule
	Config      *Config
	Name        string
	_props      *Props
	_handlers   *HandleRegistryValue
	_extensions *extensions.Extensions
	_logger     logs.LogAgent // 日志代理接口
	_stopper    chan struct{} // 是否已停止
}

func (ns *NetworkSystem) Open(addr string) error {
	return nil
}

func (ns *NetworkSystem) Logger() logs.LogAgent {
	return ns._logger
}

func (ns *NetworkSystem) IsStopped() bool {
	select {
	case <-ns._stopper:
		return true
	default:
		return false
	}
}

/*func (ns *NetworkSystem) spawnNamed(name string) (*ClientId, error) {
	var (
		cid *ClientId
		err error
	)
}*/
