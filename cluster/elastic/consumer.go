package elastic

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/network/connectPool"
)

// ElastcConsumer 可关闭的带任务处理器的消费者
type ElastcConsumer struct {
    pool  *connectPool.ConnectPool
	client   sarama.ConsumerGroup
	ctx      context.Context
    cancel context.CancelFunc
	topics    []string
}

// NewElastcConsumer 构造

func NewElastcConsumer(cfg *ElasticConfig) *ElastcConsumer {


	config := sarama.NewConfig()

	config.Version = sarama.V2_8_0_0
	config.Consumer.Return.Errors = false
	config.Consumer.Offsets.AutoCommit.Interval = time.Second * 1

    cli, err := sarama.NewConsumerGroup(strings.Split(cfg.ElasticKafka.Broker, ","), cfg.ElasticKafka.Group, config)
    if err != nil{
        return nil
    }
	res := &ElastcConsumer{}

	res.topics = cfg.ElasticKafka.Topics

    res.pool = connectPool.NewConnectPool(cfg.ElasticAddress, connectPool.IdleConfig{
        NewConn: NewElasticConsumerConnect,
    })
    res.ctx,res.cancel = context.WithCancel(context.Background())
    res.client = cli
	return res
}
// Setup 启动
func (ec *ElastcConsumer) Setup(s sarama.ConsumerGroupSession) error {
	log.Printf("[main] consumer.Setup memberID=[%s]", s.MemberID())
	return nil
}

// Cleanup 当退出时
func (ec *ElastcConsumer) Cleanup(s sarama.ConsumerGroupSession) error {
	log.Printf("[main] consumer.Cleanup memberID=[%s]", s.MemberID())
	return nil
}

// ConsumeClaim 消费日志
func (ec *ElastcConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			for _,v := range ec.topics{
				if v == message.Topic{
					msg := &prvs.PostLogsMessage{
						Vaddr: string(message.Key),
						Msg: message.Value,
					}
					ec.pool.RequestMessage(msg,0)
					session.MarkMessage(message, "")
					break
				}
			}
		case <-ec.ctx.Done():
			return nil
		}
	}
}



func (ec *ElastcConsumer) Start() error {
	go func() {
		for {
			select {
			case <-ec.ctx.Done():
				return
			default:
				ec.client.Consume(ec.ctx, ec.topics, ec)
			}
		}
	}()
    return nil
}

func (e *ElastcConsumer) Shudown(){
    if e.pool != nil{
        e.pool.Shudown()
    }
    e.cancel()
}