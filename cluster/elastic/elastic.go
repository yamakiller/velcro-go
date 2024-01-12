package elastic

import (
	"context"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/olivere/elastic"
)

const (
	elasticVelgroGoLogTopic = "velgro_go_logs"
	elasticVelgroGoLogGroup = "velgro_go"
)

// ElastcConsumer 可关闭的带任务处理器的消费者
type Elastc struct {
	topic string

	ecli   *elastic.Client
	kcli   sarama.ConsumerGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// NewElastcConsumer 构造

func NewElastc(taddr string, saddr string) *Elastc {

	config := sarama.NewConfig()

	config.Version = sarama.V2_8_0_0
	config.Consumer.Return.Errors = false
	config.Consumer.Offsets.AutoCommit.Interval = time.Second * 1

	kcli, err := sarama.NewConsumerGroup([]string{saddr}, elasticVelgroGoLogGroup, config)
	if err != nil {
		return nil
	}
	ecli, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(taddr))
	if err != nil {
		return nil
	} else {
		// ping检查
		if _, _, err := ecli.Ping(taddr).Do(context.Background()); nil != err {
			return nil
		}
		defer ecli.Stop()
	}

	res := &Elastc{
		topic: elasticVelgroGoLogTopic,
		ecli:  ecli,
		kcli:  kcli,
	}
	res.ctx, res.cancel = context.WithCancel(context.Background())
	return res
}

// Setup 启动
func (ec *Elastc) Setup(s sarama.ConsumerGroupSession) error {
	log.Printf("[main] consumer.Setup memberID=[%s]", s.MemberID())
	return nil
}

// Cleanup 当退出时
func (ec *Elastc) Cleanup(s sarama.ConsumerGroupSession) error {
	log.Printf("[main] consumer.Cleanup memberID=[%s]", s.MemberID())
	return nil
}

// ConsumeClaim 消费日志
func (ec *Elastc) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			if message.Topic == ec.topic {
				js := map[string]string{
					string(message.Key): string(message.Value),
				}
				ec.ecli.Index().Index(string(message.Key)).Type("doc").BodyJson(js).Do(context.Background())
			}

		case <-ec.ctx.Done():
			return nil
		}
	}
}

func (ec *Elastc) Start() error {
	go func() {
		for {
			select {
			case <-ec.ctx.Done():
				return
			default:
				ec.kcli.Consume(ec.ctx, []string{ec.topic}, ec)
			}
		}
	}()
	return nil
}

func (e *Elastc) Shudown() {
	e.cancel()

	if e.kcli != nil {
		e.kcli.Close()
	}
}
