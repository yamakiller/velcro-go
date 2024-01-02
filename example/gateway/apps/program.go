package apps

import (
	"os"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/cluster/gateway"
	"github.com/yamakiller/velcro-go/vlog"
)

type Program struct {
	System *gateway.Gateway
}

func (p *Program) Start(s service.Service) error {

	p.System = gateway.New(gateway.WithLAddr("127.0.0.1:8800"), gateway.WithVAddr("127.0.0.1:8810"))

	if err := p.System.Start(); err != nil {
		vlog.Errorf("Listening 127.0.0.1:8800 fail[error:%s]", err.Error())
		return err
	}

	vlog.Info("Listening 127.0.0.1:8800")

	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.System != nil {
		vlog.Info("Service Shutdown")
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
