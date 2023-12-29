package apps

import (
	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/configs"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/utils/files"
)

var appName = "login.service"

type Program struct {
	service  *loginService
	logAgent logs.LogAgent
}

func (p *Program) Start(s service.Service) error {
	p.logAgent = serve.ProduceLogger("login")
	p.logAgent.Info("[PROGRAM]", "LoginService Start loading environment variables")

	envs.With(&envs.YAMLEnv{})
	if err := envs.Instance().Load("config", files.NewLocalPathFull("config.yaml"), &configs.Config{}); err != nil {
		p.logAgent.Fatal("[PROGRAM]", "Failed to load environment variables[error:%s]", err.Error())
		p.logAgent.Close()
		return err
	}

	p.service = &loginService{}
	if err := p.service.Start(p.logAgent); err != nil {
		p.logAgent.Info("[PROGRAM]", "LoginService Failed to start network service, %s", err.Error())
		return err
	}
	p.logAgent.Info("[PROGRAM]", "LoginService Start network service completed")

	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.service != nil {
		p.logAgent.Info("[PROGRAM]", "LoginService service terminating")
		p.service.Stop()
		p.service = nil

		p.logAgent.Info("[PROGRAM]", "LoginService service terminated")
		p.logAgent.Close()
		p.logAgent = nil
	}
	return nil
}
