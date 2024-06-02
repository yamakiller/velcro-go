package apps

import (
	// "strings"

	"github.com/kardianos/service"
	// "github.com/yamakiller/velcro-go/cluster/logs"

	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/gateway.service/configs"
	"github.com/yamakiller/velcro-go/utils/files"
	"github.com/yamakiller/velcro-go/vlog"
)

type Program struct {
	service *gatewayService
}

func (p *Program) Start(s service.Service) error {
	if err := vlog.SetLogFile("", "GatewayService"); err != nil {
		return err
	}
	vlog.Info("[PROGRAM]", "Gateway Start loading environment variables")

	envs.With(&envs.YAMLEnv{})
	if err := envs.Instance().Load("configs",
		files.NewLocalPathFull("config.yaml"),
		&configs.Config{}); err != nil {
		vlog.Fatal("[PROGRAM]", "Failed to load environment variables", err)

		return err
	}

	// vaddr := strings.ReplaceAll(strings.ToLower("gateway@"+envs.Instance().Get("configs").(*configs.Config).Server.VAddr), ":", ".")
	// vlog.SetOutput(logs.NewElastic(envs.Instance().Get("configs").(*configs.Config).LogRemoteAddr, vaddr))

	vlog.Info("[PROGRAM]", "Gateway Loading environment variables is completed")
	vlog.Info("[PROGRAM]", "Gateway Start the network service")
	p.service = &gatewayService{}
	if err := p.service.Start(); err != nil {
		vlog.Info("[PROGRAM]", "Gateway Failed to start network service", err)
		return err
	}
	vlog.Info("[PROGRAM]", "Gateway Start network service completed ", envs.Instance().Get("configs").(*configs.Config).Server.LAddr)

	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.service != nil {
		vlog.Info("[PROGRAM]", "Gateway service terminating")
		p.service.Stop()
		p.service = nil
		vlog.Info("[PROGRAM]", "Gateway service terminated")
	}
	return nil
}
