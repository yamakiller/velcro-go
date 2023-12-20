package classes

import (
	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/network"
)

func New(interAddress string, // inter 服务地址
	interSystem *network.NetworkSystem, // inter 服务系统接口
	group ClientGroup, // 连接者管理组
	encryption *Encryption, // 通信加密组
) *Gateway {
	return &Gateway{listenAddress: interAddress,
		serverSystem: interSystem,
		clientGroup:  group,
		encryption:   encryption}
}

// Gateway 网关
type Gateway struct {
	virtualAddress string // 网关局域网虚拟地址
	listenAddress  string
	serverSystem   *network.NetworkSystem
	clientGroup    ClientGroup
	encryption     *Encryption
	routeGroup     *router.RouterGroup
}

func (dg *Gateway) Start() error {
	//
	//router.Loader()

	if err := dg.serverSystem.Open(dg.listenAddress); err != nil {
		return err
	}

	return nil
}

func (dg *Gateway) Stop() error {
	if dg.serverSystem != nil {
		dg.serverSystem.Info("Gateway Shutdown Closing linker")
		dg.clientGroup.Clear()
		dg.serverSystem.Info("Gateway Shutdown Closed linker")
		dg.serverSystem.Shutdown()
		dg.serverSystem.Info("Gateway Shutdown")
		dg.serverSystem = nil
	}
	return nil
}
