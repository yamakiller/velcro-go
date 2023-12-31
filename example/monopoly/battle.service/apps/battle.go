package apps

import (
	"github.com/yamakiller/velcro-go/cluster/serve"
	"github.com/yamakiller/velcro-go/example/monopoly/login.service/dba/rds"
	"github.com/yamakiller/velcro-go/logs"
)

type battleService struct {
	battle *serve.Servant
}

func (bs *battleService) Start(logAgent logs.LogAgent) error {
	return nil
}

func (bs *battleService) Stop() error {

	if bs.battle != nil {
		bs.battle.Stop()
		bs.battle = nil
	}

	rds.Disconnect()

	return nil
}
