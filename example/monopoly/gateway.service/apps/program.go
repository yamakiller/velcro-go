package apps

import (
	"github.com/yamakiller/velcro-go/cluster/gateway"
	"github.com/yamakiller/velcro-go/cluster/service"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/gateway.service/configs"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/utils/files"
)

type Program struct {
	service  *gatewayService
	logAgent logs.LogAgent
}

func (p *Program) Start(s service.Service) error {

	p.logAgent = gateway.ProduceLogger()

	p.logAgent.Info("[PROGRAM]", "Gateway Start loading environment variables")
	if err := envs.Instance().Load("configs",
		files.NewLocalPathFull("config.yaml"),
		&configs.Config{}); err != nil {
		p.logAgent.Fatal("[PROGRAM]", "Failed to load environment variables[error:%s]", err.Error())
		p.logAgent.Close()
		return err
	}
	p.logAgent.Info("[PROGRAM]", "Gateway Loading environment variables is completed")
	p.logAgent.Info("[PROGRAM]", "Gateway Start the network service")
	p.service = &gatewayService{}
	if err := p.service.Start(p.logAgent); err != nil {
		p.logAgent.Info("[PROGRAM]", "Gateway Failed to start network service, %s", err.Error())
		return err
	}
	p.logAgent.Info("[PROGRAM]", "Gateway Start network service completed")
	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.service != nil {
		p.logAgent.Info("[PROGRAM]", "Gateway service terminating")
		p.service.Stop()
		p.service = nil
		p.logAgent.Info("[PROGRAM]", "Gateway service terminated")
		p.logAgent.Close()
		p.logAgent = nil
	}
	return nil
}
