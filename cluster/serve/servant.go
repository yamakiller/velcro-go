package serve

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/yamakiller/velcro-go/cluster/repeat"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/syncx"
	"google.golang.org/protobuf/proto"
)

func New(options ...ConnConfigOption) *Servant {
	config := configure(options...)
	s := &Servant{producer: config.Producer,
		clients: make(map[network.CIDKEY]*network.ClientID),
		vaddrs:  newSliceMap()}
	s.NetworkSystem = network.NewTCPSyncServerNetworkSystem(
		network.WithMetricProviders(config.MetricsProvider),
		network.WithKleepalive(config.Kleepalive),
		network.WithProducer(s.spawConn))
	return s
}

type Servant struct {
	*network.NetworkSystem
	producer    func(*ServantClientConn) ServantClientActor
	clients     map[network.CIDKEY]*network.ClientID
	clientMutex sync.Mutex
	vaddrs      *sliceMap // 虚地址表
}

func (s *Servant) Open(addr string) error {
	return s.NetworkSystem.Open(addr)
}

func (s *Servant) Shutdown() {
	s.clientMutex.Lock()
	for _, clientId := range s.clients {
		clientId.UserClose()
	}
	s.clientMutex.Unlock()
	s.NetworkSystem.Shutdown()
}

func (s *Servant) PostMessage(addr string, msg proto.Message) error {
	b, err := messages.MarshalMessageProtobuf(0, msg)
	if err != nil {
		panic(err)
	}

	var resultErr error = nil
	repeat.Repeat(repeat.FnWithCounter(func(n int) error {
		if n >= 3 {
			return nil
		}

		bucket := s.vaddrs.getBucket(addr)
		clientId, ok := bucket.Get(addr)
		if !ok {
			resultErr = fmt.Errorf("post addr %s %s message unfound target", addr, reflect.TypeOf(msg))
			return resultErr
		}

		if err := clientId.(*network.ClientID).PostUserMessage(b); err != nil {
			resultErr = err
			return resultErr
		}

		resultErr = nil
		return nil
	}),
		repeat.StopOnSuccess(),
		repeat.WithDelay(repeat.ExponentialBackoff(500*time.Millisecond).Set()))
	return resultErr
}

func (s *Servant) spawConn(system *network.NetworkSystem) network.Client {
	return &ServantClientConn{
		Servant: s,
		recvice: circbuf.New(32768, &syncx.NoMutex{}),
	}
}
