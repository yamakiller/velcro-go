package apps

import (
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/client"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/configs"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	"github.com/yamakiller/velcro-go/logs"
)
var appName = "login.service"

type Program struct {
	System   *client.LoginService
	logAgent *logs.DefaultAgent
}

func (p *Program) Start(s service.Service) error {
	logLevel := logrus.DebugLevel
	if os.Getenv("DEBUG") != "" {
		logLevel = logrus.InfoLevel
	}

	logDir := p.getDirLog()

	pLogHandle := logs.SpawnFileLogrus(logLevel, logDir, "")
	p.logAgent = &logs.DefaultAgent{}
	p.logAgent.WithHandle(pLogHandle)

	cfgFilePath, err := configs.GetLocalConfigFilePath()
	if err != nil {
		p.logAgent.Error(appName, "%s", err.Error())
		p.logAgent.Close()
		return err
	}

	attr := configs.Config{}
	envs.With(&envs.YAMLEnv{})
	if err := envs.Instance().Load("config", cfgFilePath, &attr); err != nil {
		p.logAgent.Error(appName, "Load %s config file fail-%s", cfgFilePath, err.Error())
		p.logAgent.Close()
		return err
	}



	//1.连接-Redis
	rds.WithAddr(attr.Redis.Addr)
	rds.WithPwd(attr.Redis.Pwd)
	rds.WithDialTimeout(attr.Redis.DialTimeout)
	rds.WithReadTimeout(attr.Redis.ReadTimeout)
	rds.WithWriteTimeout(attr.Redis.WriteTimeout)
	if err := rds.Connection(); err != nil {
		p.logAgent.Error(appName, "connection %v redis fail-%s", attr.Redis.Addr, err.Error())
		p.logAgent.Close()
		return err
	}

	p.logAgent.Info(appName, "redis is connected")

	p.System = client.New()
	if err := p.System.Open("127.0.0.1:8860"); err != nil {
		p.System.Error("Listening 127.0.0.1:8860 fail[error:%s]", err.Error())
		return err
	}

	p.System.Info("Listening 127.0.0.1:8860")

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
