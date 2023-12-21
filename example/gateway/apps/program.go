package apps

import (
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	"github.com/yamakiller/velcro-go/cluster/gateway"
	"github.com/yamakiller/velcro-go/logs"
)

type Program struct {
	System   *gateway.Gateway
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

	p.System = gateway.New( gateway.WithLoggerAgent(p.logAgent),gateway.WithLAddr("127.0.0.1:8800"),gateway.WithVAddr("127.0.0.1:8810"),gateway.WithRouterURI("./routes.yml"))

	if err := p.System.Start(); err != nil {
		p.logAgent.Error("","Listening 127.0.0.1:8800 fail[error:%s]", err.Error())
		return err
	}

	p.logAgent.Info("","Listening 127.0.0.1:8800")

	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.System != nil {
		p.logAgent.Info("","Service Shutdown")
		p.System.Stop()
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
