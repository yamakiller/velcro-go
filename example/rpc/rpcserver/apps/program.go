package apps

import (
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/rpc/server"
)

type Program struct {
	_s       *server.RpcServer
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

	p._s = server.New()
	p._s.WithPool(NewClientPools(p._s))
	if err := p._s.Open("127.0.0.1:9870"); err != nil {
		p._s.Error("Listening 127.0.0.1:9870 fail[error:%s]", err.Error())
		return err
	}
	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p._s != nil {
		p._s.Info("Service Shutdown")
		p._s.Shutdown()
		p._s = nil
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
