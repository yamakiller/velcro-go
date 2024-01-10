package elastic

import (
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/network/connectPool"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func NewElasticProducerConnect(options ...connectPool.ConnConfigOption) connectPool.IConnect {
	res := &ElasticProducerConnect{}
	return res
}

type ElasticProducerConnect struct {
	connectPool.BaseConnect
	client sarama.SyncProducer
}

func (ep *ElasticProducerConnect) Timeout() int64 {
	panic(nil)
}
func (ep *ElasticProducerConnect) Dial(address string,timeout time.Duration) error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.NoResponse                                  // Only wait for the leader to ack
	config.Producer.Flush.Frequency =500 * time.Millisecond // Flush batches every 500ms
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	urls := strings.Split(address, ",")
	if cli, err := sarama.NewSyncProducer(urls,config); err != nil {
		return err
	} else {
		ep.client = cli
	}
	
	return nil
}
func (ep *ElasticProducerConnect) Redial() error {
	panic(nil)
}

func (ep *ElasticProducerConnect) RequestMessage(message protoreflect.ProtoMessage,timeout int64) (connectPool.IFuture, error) {
	mm := message.(*prvs.PostLogsMessage)
	msg := &sarama.ProducerMessage{
        Topic: mm.Topic,
        Key:   sarama.StringEncoder(mm.Vaddr),
        Value: sarama.ByteEncoder(mm.Msg),
    }
	ep.client.SendMessage(msg)
	return nil,nil
}

func (ep *ElasticProducerConnect) Close() {
	if ep.client != nil{
		ep.client.Close()
	}
}