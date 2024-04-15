package apps

import (
	"github.com/kardianos/service"
	"github.com/yamakiller/velcro-go/example/behavior/client"
	"github.com/yamakiller/velcro-go/example/behavior/kafka"
)

type Program struct {
	// System *network.NetworkSystem
}

func (p *Program) Start(s service.Service) error {

	kafka.ListenAndServe("127.0.0.1:9092")
	
	client.BInt("D:/GOPATH/src/fenzhi/velcro-go/tools/behavior/test2.b3")
	// if err := p.System.Open("127.0.0.1:9860"); err != nil {
	// 	vlog.Errorf("Listening 127.0.0.1:9860 fail[error:%s]", err.Error())
	// 	return err
	// }

	// vlog.Infof("Listening 127.0.0.1:9860")

	return nil
}

func (p *Program) Stop(s service.Service) error {
	// if p.System != nil {
	// 	vlog.Info("Service Shutdown")
	// 	p.System.Shutdown()
	// 	p.System = nil
	// }
	return nil
}

func (p *Program) testStop() {
	// if p.System != nil {
	// 	vlog.Info("Service Shutdown")
	// 	p.System.Shutdown()
	// 	p.System = nil
	// }
}
