package apps

import (
	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/configs"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/utils/files"
)

type Program struct {
	service  *battleService
	logAgent logs.LogAgent
}

func (p *Program) Start(s service.Service) error {
	p.logAgent = serve.ProduceLogger("battle")
	p.logAgent.Info("[PROGRAM]", "BattleService Start loading environment variables")

	envs.With(&envs.YAMLEnv{})
	if err := envs.Instance().Load("config", files.NewLocalPathFull("config.yaml"), &configs.Config{}); err != nil {
		p.logAgent.Fatal("[PROGRAM]", "Failed to load environment variables[error:%s]", err.Error())
		p.logAgent.Close()
		return err
	}

	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.service != nil {
		p.logAgent.Info("[PROGRAM]", "BattleService service terminating")
		p.service.Stop()
		p.service = nil

		p.logAgent.Info("[PROGRAM]", "BattleService service terminated")
		p.logAgent.Close()
		p.logAgent = nil
	}
	return nil
}
