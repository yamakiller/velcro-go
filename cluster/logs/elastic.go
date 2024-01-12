package logs

import (
	"strings"
	// "time"

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

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.NoResponse                                  // Only wait for the leader to ack
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	cli, err := sarama.NewAsyncProducer([]string{addr},config)
	if err != nil {
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
	msg := &sarama.ProducerMessage{
        Topic: e.topic,
        Key:   sarama.StringEncoder(e.vaddr),
        Value: sarama.ByteEncoder(in),
    }
	e.client.Input() <- msg

	return 0,nil
}