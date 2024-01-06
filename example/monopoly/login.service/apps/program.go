package apps

import (
	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/configs"
	"github.com/yamakiller/velcro-go/utils/files"
	"github.com/yamakiller/velcro-go/vlog"
)

type Program struct {
	service *loginService
}

func (p *Program) Start(s service.Service) error {

	vlog.Info("[PROGRAM]", "LoginService Start loading environment variables")

	envs.With(&envs.YAMLEnv{})
	if err := envs.Instance().Load("config", files.NewLocalPathFull("config.yaml"), &configs.Config{}); err != nil {
		vlog.Fatal("[PROGRAM]", "Failed to load environment variables", err)
		return err
	}

	p.service = &loginService{}
	if err := p.service.Start(); err != nil {
		vlog.Info("[PROGRAM]", "LoginService Failed to start network service", err)
		return err
	}
	vlog.Info("[PROGRAM]", "LoginService Start network service completed")

	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.service != nil {
		vlog.Info("[PROGRAM]", "LoginService service terminating")
		p.service.Stop()
		p.service = nil

		vlog.Info("[PROGRAM]", "LoginService service terminated")
	}
	return nil
}
