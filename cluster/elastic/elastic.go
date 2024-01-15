package elastic

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/olivere/elastic"
	"log"
	"os"
	"time"
)

const (
	elasticVelgroGoLogTopic = "velgro_go_logs"
	elasticVelgroGoLogGroup = "velgro_go"
)

// ElastcConsumer 可关闭的带任务处理器的消费者
type Elastc struct {
	topics []string

	ecli   *elastic.Client
	kcli   sarama.ConsumerGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// NewElastcConsumer 构造

func NewElastc(taddr string, saddr string) *Elastc {

	cfg := sarama.NewConfig()

	cfg.Version = sarama.V2_8_0_0
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Offsets.Retry.Max = 3
	cfg.Consumer.Offsets.AutoCommit.Enable = true              // 开启自动提交，需要手动调用MarkMessage才有效
	cfg.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second // 间隔
	kcli, err := sarama.NewConsumerGroup([]string{saddr}, elasticVelgroGoLogGroup, cfg)
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
		topics: []string{elasticVelgroGoLogTopic},
		ecli:   ecli,
		kcli:   kcli,
	}
	res.ctx, res.cancel = context.WithCancel(context.Background())
	return res
}

// Setup 启动
func (ec *Elastc) Setup(s sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 当退出时
func (ec *Elastc) Cleanup(s sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 消费日志
func (ec *Elastc) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for message := range claim.Messages() {
		fmt.Fprintln(os.Stderr, message.Topic, "   ", string(message.Value))
		if message.Topic == elasticVelgroGoLogTopic {
			js := map[string]string{
				string(message.Key): string(message.Value),
			}
			ec.ecli.Index().Index(string(message.Key)).Type("doc").BodyJson(js).Do(context.Background())
		}
		// 处理消息成功后标记为处理, 然后会自动提交
		session.MarkMessage(message, "")
	}
	return nil
}

func (ec *Elastc) Start() error {
	go func() {
		for {
			select {
			case <-ec.ctx.Done():
				ec.kcli.Close()
				log.Println("EventConsumer ctx done")
				return
			default:
				if err := ec.kcli.Consume(ec.ctx, ec.topics, ec); err != nil {
					log.Println("EventConsumer Consume failed err is ", err.Error())
				}
			}
		}
	}()

	return nil
}

func (e *Elastc) Shudown() {
	e.cancel()
}
