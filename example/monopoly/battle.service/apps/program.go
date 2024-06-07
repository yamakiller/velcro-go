package apps

import (
	// "strings"

	"github.com/kardianos/service"
	// "github.com/yamakiller/velcro-go/cluster/logs"

	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/battle.service/configs"
	"github.com/yamakiller/velcro-go/utils/files"
	"github.com/yamakiller/velcro-go/vlog"
)

type Program struct {
	service *battleService
}

func (p *Program) Start(s service.Service) error {
	if err := vlog.SetLogFile("", "BattleService"); err != nil {
		return err
	}
	vlog.Info("[PROGRAM]", "BattleService Start loading environment variables")

	envs.With(&envs.YAMLEnv{})
	if err := envs.Instance().Load("configs", files.NewLocalPathFull("config.yaml"), &configs.Config{}); err != nil {
		vlog.Fatal("[PROGRAM]", "Failed to load environment variables", err)
		return err
	}
	// vaddr := strings.ReplaceAll(strings.ToLower("battle@"+envs.Instance().Get("configs").(*configs.Config).Server.VAddr), ":", ".")
	// vlog.SetOutput(logs.NewElastic(envs.Instance().Get("configs").(*configs.Config).LogRemoteAddr, vaddr))

	p.service = &battleService{}
	if err := p.service.Start(); err != nil {
		vlog.Fatal("[PROGRAM]", "Failed to load environment variables", err)
		return err
	}
	vlog.Info("[PROGRAM]", "BattleService Start Successfully ", envs.Instance().Get("configs").(*configs.Config).Server.LAddr)

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
