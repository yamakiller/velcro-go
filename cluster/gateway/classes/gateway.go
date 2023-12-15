package classes

import (
	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/network"
)

func New(interAddress string, // inter 服务地址
	interSystem *network.NetworkSystem, // inter 服务系统接口
	group GatewayLinkerGroup, // 连接者管理组
	encryption *Encryption, // 通信加密组
) *Gateway {
	return &Gateway{intelServerAddress: interAddress,
		intelServerSystem: interSystem,
		linkerGroup:       group,
		encryption:        encryption}
}

// Gateway 网关
type Gateway struct {
	intelServerAddress string
	intelServerSystem  *network.NetworkSystem
	linkerGroup        GatewayLinkerGroup
	encryption         *Encryption
}

func (dg *Gateway) Start(s service.Service) error {

	if err := dg.intelServerSystem.Open(dg.intelServerAddress); err != nil {
		return err
	}

	return nil
}

func (dg *Gateway) Stop(s service.Service) error {
	if dg.intelServerSystem != nil {
		dg.intelServerSystem.Info("Gateway Shutdown Closing linker")
		dg.linkerGroup.Clear()
		dg.intelServerSystem.Info("Gateway Shutdown Closed linker")
		dg.intelServerSystem.Shutdown()
		dg.intelServerSystem.Info("Gateway Shutdown")
		dg.intelServerSystem = nil
	}
	return nil
}
