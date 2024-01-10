package elastic

import (
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/network/connectPool"
)


func NewElasticProducer(cfg *ElasticConfig) *ElasticProducer {
	
	pool :=connectPool.NewConnectPool(cfg.ElasticKafka.Broker,connectPool.IdleConfig{
		NewConn: NewElasticProducerConnect,
	})
	res := &ElasticProducer{
		cfg: cfg,
		pool: pool,
	}
	return res
}


type ElasticProducer struct {
	cfg *ElasticConfig
	pool *connectPool.ConnectPool
}

func (ep *ElasticProducer) PostMessage(vaddr string, message []byte) error {
	for _, topic := range ep.cfg.ElasticKafka.Topics{
		msg := &prvs.PostLogsMessage{
			Topic: topic,
			Vaddr: vaddr,
			Msg: message,
		}
		ep.pool.RequestMessage(msg,0)
	}
	return nil
}


