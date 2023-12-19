package apps

import (
	"os"
	"path/filepath"
	"time"

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/rpc/protos"
	"github.com/yamakiller/velcro-go/example/rpc/rpcclient/configs"
	"github.com/yamakiller/velcro-go/example/rpc/rpcclient/internet"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/rpc/rpcclient"
)

var appName string = "test-rpc-client"

type Program struct {
	_c       *rpcclient.Conn
	logAgent *logs.DefaultAgent
}

func (p *Program) Start(s service.Service) error {
	logLevel := logrus.DebugLevel
	if os.Getenv("DEBUG") != "" {
		logLevel = logrus.InfoLevel
	}

	pLogHandle := logs.SpawnFileLogrus(logLevel, "", "")
	p.logAgent = &logs.DefaultAgent{}
	p.logAgent.WithHandle(pLogHandle)

	cfgFilePath, err := p.GetLocalConfigFilePath()
	if err != nil {
		p.logAgent.Error(appName, "load config fail:[error:%s]", err.Error())
		p.logAgent.Close()
		return err
	}

	config := configs.Config{}
	envs.With(config.IEnv())
	if err := envs.Instance().Load("config", cfgFilePath, &config); err != nil {
		p.logAgent.Error(appName, "Load %s config file fail-%s", cfgFilePath, err.Error())
		p.logAgent.Close()
		return err
	}
	p._c = rpcclient.NewConn(
		rpcclient.WithClosed(internet.Closed),
		rpcclient.WithReceive(internet.Receive))
	if err := p._c.Dial(config.TargetAddr, time.Duration(4)*time.Millisecond); err != nil {
		p.logAgent.Error("Dial %s fail[error:%s]", config.TargetAddr, err.Error())
		return err
	}

	if msg, err := p._c.RequestMessage(&protos.Auth{Msg: "123456"}, 30000); err == nil {
		p.logAgent.Info("Auth %s success", msg.(*protos.Auth).Msg)
	} else {
		p.logAgent.Error("RequestMessage fail[error:%s]", err.Error())
	}
	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p._c != nil {
		p._c.Close()
		p._c = nil
	}
	return nil
}
func (p *Program) GetLocalConfigFilePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	exPath := filepath.Dir(ex)
	cfgFilePath := filepath.Join(exPath, "config.yml")

	return cfgFilePath, nil
}
