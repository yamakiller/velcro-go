package apps

import (
	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/configs"
	"github.com/yamakiller/velcro-go/utils/files"
	"github.com/yamakiller/velcro-go/vlog"
)

type Program struct {
	service *battleService
}

func (p *Program) Start(s service.Service) error {
	vlog.Info("[PROGRAM]", "BattleService Start loading environment variables")

	envs.With(&envs.YAMLEnv{})
	if err := envs.Instance().Load("config", files.NewLocalPathFull("config.yaml"), &configs.Config{}); err != nil {
		vlog.Fatal("[PROGRAM]", "Failed to load environment variables", err.Error())
		return err
	}
	p.service = &battleService{}
	if err := p.service.Start(p.logAgent); err != nil {
		p.logAgent.Fatal("[PROGRAM]", "Failed to load environment variables[error:%s]", err.Error())
		p.logAgent.Close()
		return err
	}
	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.service != nil {
		vlog.Info("[PROGRAM]", "BattleService service terminating")
		p.service.Stop()
		p.service = nil

		vlog.Info("[PROGRAM]", "BattleService service terminated")
	}
	return nil
}
