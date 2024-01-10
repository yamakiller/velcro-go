package apps

import (
	"github.com/yamakiller/velcro-go/cluster/elastic"
	"github.com/yamakiller/velcro-go/envs"
	"github.com/yamakiller/velcro-go/example/monopoly/logs.service/configs"
)

type logsService struct {
	logs * elastic.ElastcConsumer
}

func (ls *logsService) Start() error {
	ls.logs =  elastic.NewElastcConsumer(&envs.Instance().Get("configs").(*configs.Config).Elastic)
	if ls.logs == nil {
		return nil
	}
	if err := ls.logs.Start(); err != nil {
		return err
	}
	return nil
}


func (ls *logsService) Stop() error {
	if ls.logs != nil {
		ls.logs.Shudown()
	}
	
	return nil
}
