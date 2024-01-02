package apps

import (
	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/rpc/server"
	"github.com/yamakiller/velcro-go/vlog"
)

type Program struct {
	service *server.RpcServer
}

func (p *Program) Start(s service.Service) error {

	p.service = server.New()
	p.service.WithPool(NewClientPools(p.service))
	if err := p.service.Open("127.0.0.1:9870"); err != nil {
		vlog.Errorf("Listening 127.0.0.1:9870 fail[error:%s]", err.Error())
		return err
	}
	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.service != nil {
		vlog.Info("Service Shutdown")
		p.service.Shutdown()
		p.service = nil
	}
	return nil
}
