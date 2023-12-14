package apps

import (
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	"github.com/yamakiller/velcro-go/example/tcpserver/uclient"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/network"
)

type Program struct {
	System   *network.NetworkSystem
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

	p.System = network.NewTCPNetworkSystem(network.WithMetricProviders(testPrometheusProvider(8091)),
		network.WithLoggerFactory(func(system *network.NetworkSystem) logs.LogAgent {

			return p.logAgent
		}),
		network.WithProducer(uclient.NewTestClient))

	if err := p.System.Open("127.0.0.1:9860"); err != nil {
		p.System.Error("Listening 127.0.0.1:9860 fail[error:%s]", err.Error())
		return err
	}

	p.System.Info("Listening 127.0.0.1:9860")

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

func (p *Program) testStop() {
	if p.System != nil {
		p.System.Info("Service Shutdown")
		p.System.Shutdown()
		p.System = nil
	}
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
