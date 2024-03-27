package kafka

import (
	"context"
	// "sync"
	"github.com/IBM/sarama"
)

var (
	_topic    string = "behavior_node"
	_msg      chan ProduceMsg
	_producer sarama.SyncProducer
	_ctx      context.Context
	_err      error
	// _m        sync.Mutex
)

type ProduceMsg struct {
	Key string
	Msg string
}

func WriteCurrTree(tree string,node string) error {



	if _msg != nil {
		msg := ProduceMsg{
			Key: "test",
			Msg: tree +","+node,
		}
		_msg <- msg
	}
	return nil
}
func ListenAndServe(addr string) error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewHashPartitioner
	_ctx = context.Background()
	_msg = make(chan ProduceMsg, 128)
	_producer, _err = sarama.NewSyncProducer([]string{addr}, config)
	if _err != nil {
		return _err
	}
	go serve()
	return nil
}
func serve() {
	defer func() {
		_ = _producer.Close()
	}()
	for {
		select {
		case <-_ctx.Done():
			return
		case promsg, ok := <-_msg:
			if ok {
				sendMsg(promsg)
			}
		}
	}
}
func sendMsg(promsg ProduceMsg) error {

	msg := &sarama.ProducerMessage{
		Topic: _topic,
		Value: sarama.StringEncoder(promsg.Msg),
		Key:   sarama.StringEncoder(promsg.Key),
	}
	_, _, err := _producer.SendMessage(msg)
	if err != nil {
		return err
	}
	// _log.Info("", "partition :%+v  offset:%+v ", partition, offset)
	return nil
}
