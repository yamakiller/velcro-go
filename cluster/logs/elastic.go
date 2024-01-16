package logs

import (
	"strings"
	"time"

	"github.com/IBM/sarama"
)

const(
	elasticVelgroGoLogTopic = "velgro_go_logs"
)

// Elastic  vaddr 用于识别来源
func NewElastic(addr, vaddr string) *Elastic {
	if addr == "" || vaddr == "" {
		panic("addr or vaddr is nil")
	}
	if !strings.EqualFold(vaddr, strings.ToLower(vaddr)){
		panic("vaddr string uppercase")
	}
	
	if strings.Contains(vaddr, ":") {
		panic("vaddr have ':'")
	}

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0
	cfg.Producer.RequiredAcks = sarama.WaitForAll // 三种模式任君选择
	cfg.Producer.Partitioner = sarama.NewHashPartitioner
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	cfg.Producer.Retry.Max = 3
	cfg.Producer.Retry.Backoff = 100 * time.Millisecond
	cli, err := sarama.NewAsyncProducer([]string{addr}, cfg)
	if err != nil{
		return nil
	}

	res := &Elastic{
		topic: elasticVelgroGoLogTopic,
		vaddr: vaddr,
		client: cli,
	}

	return res
}
type Elastic struct {
	vaddr string
	topic string
	client sarama.AsyncProducer
}

func (e *Elastic) Write(in []byte) (int,error){

	e.client.Input() <- &sarama.ProducerMessage{
		Value: sarama.ByteEncoder(in),
		Key:   sarama.StringEncoder(e.vaddr),
		Topic: e.topic,
	}

	return 0,nil
}