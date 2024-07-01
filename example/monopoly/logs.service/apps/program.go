package apps

import (
	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/logs.service/configs"
	"github.com/yamakiller/velcro-go/utils/files"
	"github.com/yamakiller/velcro-go/vlog"
)

type Program struct {
	service *logsService
}

func (p *Program) Start(s service.Service) error {
	if err := vlog.SetLogFile("", "LogsService"); err != nil {
		return err
	}
	vlog.Info("[PROGRAM]", "Logs Start loading environment variables")

	envs.With(&envs.YAMLEnv{})
	if err := envs.Instance().Load("configs",
		files.NewLocalPathFull("config.yaml"),
		&configs.Config{}); err != nil {
		vlog.Fatal("[PROGRAM]", "Failed to load environment variables", err)

		return err
	}
	vlog.Info("[PROGRAM]", "Logs Loading environment variables is completed")
	vlog.Info("[PROGRAM]", "Logs Start the network service")
	p.service = &logsService{}
	if err := p.service.Start(); err != nil {
		vlog.Info("[PROGRAM]", "Logs Failed to start network service", err)
		return err
	}
	vlog.Info("[PROGRAM]", "Logs Start network service successfully")
	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.service != nil {
		vlog.Info("[PROGRAM]", "Logs service terminating")
		p.service.Stop()
		p.service = nil
		vlog.Info("[PROGRAM]", "Logs service terminated")
	}
	return nil
}
