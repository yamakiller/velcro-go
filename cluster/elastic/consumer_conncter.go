package elastic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/olivere/elastic"
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/network/connectPool"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func NewElasticConsumerConnect(options ...connectPool.ConnConfigOption) connectPool.IConnect {
	res := &ElasticConsumerConnect{}
	return res
}

type ElasticConsumerConnect struct {
	connectPool.BaseConnect
	client *elastic.Client
}

func (ec *ElasticConsumerConnect) Timeout() int64 {
	panic(nil)
}
func (ec *ElasticConsumerConnect) Dial(address string,timeout time.Duration) error {
	urls := strings.Split(address, ",")

	if cli, err := elastic.NewClient(elastic.SetSniff(false),elastic.SetURL(urls...)); err != nil {
		return err
	} else {
		// ping检查
		if _, _, err := cli.Ping(urls[0]).Do(context.Background()); nil != err {
			return err
		}
		defer cli.Stop()
		ec.client = cli
	}
	
	return nil
}
func (ec *ElasticConsumerConnect) Redial() error {
	panic(nil)
}

func (ec *ElasticConsumerConnect) RequestMessage(message protoreflect.ProtoMessage,timeout int64) (connectPool.IFuture, error) {
	service := ec.client.Bulk()
	req := elastic.NewBulkIndexRequest().Index(message.(*prvs.PostLogsMessage).Vaddr).Type("doc").Doc(json.RawMessage(message.(*prvs.PostLogsMessage).Msg))
	service.Add(req)
	if _, err := service.Do(context.Background()); nil != err {
		return nil,err
	}
	return nil,nil
}

func (ec *ElasticConsumerConnect) Close() {

}