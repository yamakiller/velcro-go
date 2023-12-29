package apps

import (
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/accounts"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/utils/files"
)

var appName = "login.service"

type Program struct {
	system   *LoginService
	logAgent logs.LogAgent
}

func (p *Program) Start(s service.Service) error {
	p.logAgent = serve.ProduceLogger("loging")
	p.logAgent.Info("[PROGRAM]", "Gateway Start loading environment variables")

	envs.With(&envs.YAMLEnv{})
	if err := envs.Instance().Load("config", files.NewLocalPathFull("config.yaml"), &configs.Config{}); err != nil {
		p.logAgent.Fatal("[PROGRAM]", "Failed to load environment variables[error:%s]", err.Error())
		p.logAgent.Close()
		return err
	}

	//1.连接-Redis
	rds.WithAddr(cfg.Redis.Addr)
	rds.WithPwd(cfg.Redis.Pwd)
	rds.WithDialTimeout(cfg.Redis.DialTimeout)
	rds.WithReadTimeout(cfg.Redis.ReadTimeout)
	rds.WithWriteTimeout(cfg.Redis.WriteTimeout)
	if err := rds.Connection(); err != nil {
		p.logAgent.Error(appName, "connection %v redis fail-%s", cfg.Redis.Addr, err.Error())
		p.logAgent.Close()
		return err
	}

	p.logAgent.Info(appName, "redis is connected")
	if err := accounts.Init(); err != nil {
		p.logAgent.Error(appName, "accounts init fail-%s", err.Error())
		p.logAgent.Close()
		return err
	}

	p.System = new(p.logAgent)
	if err := p.System.Open(cfg.Server.LAddr); err != nil {
		p.System.Error("Listening %s fail[error:%s]", cfg.Server.LAddr, err.Error())
		return err
	}

	p.System.Info("Listening %s=>%s", cfg.Server.LAddr, cfg.Server.VAddr)

	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.System != nil {
		p.System.Info("Service Shutdown")
		p.System.Shutdown()
		p.System = nil
	}
	return nil
}

func (p *Program) getDirLog() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	exPath := filepath.Dir(ex)
	logDir := filepath.Join(exPath, "test-logs")

	if !p.isDirExits(logDir) {
		if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
			return ""
		}
	}
	return logDir
}

func (p Program) isDirExits(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
