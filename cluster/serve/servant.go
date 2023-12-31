package serve

import (
	"sync"

	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/syncx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func New(options ...ServantConfigOption) *Servant {
	config := configure(options...)
	s := &Servant{
		Config:  config,
		clients: make(map[network.CIDKEY]*network.ClientID)}
	s.System = network.NewTCPSyncServerNetworkSystem(
		network.WithLoggerFactory(func(system *network.NetworkSystem) logs.LogAgent {
			return config.Logger
		}),
		network.WithMetricProviders(config.MetricsProvider),
		network.WithKleepalive(config.Kleepalive),
		network.WithProducer(s.spawConn),
		network.WithVAddr(config.Name+"@"+config.VAddr))
	return s
}

type Servant struct {
	System      *network.NetworkSystem
	Config      *ServantConfig
	clients     map[network.CIDKEY]*network.ClientID
	clientMutex sync.Mutex
	routeGroup  *router.RouterGroup
}

func (s *Servant) Start() error {
	if s.Config.Router != nil {
		routeGroup, err := router.Loader(s.Config.Router.URI,
			router.WithAlgorithm(s.Config.Router.Algorithm),
			router.WithDialTimeout(s.Config.Router.DialTimeout),
			router.WithKleepalive(s.Config.Router.Kleepalive))
		if err != nil {
			return nil
		}

		s.routeGroup = routeGroup
		s.routeGroup.Open()
	}

	if err := s.System.Open(s.Config.LAddr); err != nil {
		s.routeGroup.Shutdown()
		return err
	}

	return nil
}

func (s *Servant) Stop() error {
	if s.routeGroup != nil {
		s.routeGroup.Shutdown()
	}

	if s.System != nil {
		s.System.Info(s.Config.Name + " Shutdown Closing client")
		s.clientMutex.Lock()
		for _, clientId := range s.clients {
			clientId.UserClose()
		}
		s.clientMutex.Unlock()
		s.System.Info(s.Config.Name + " Shutdown Closed client")
		s.System.Shutdown()
		s.System.Info(s.Config.Name + " Shutdown")
		s.System = nil
	}

	s.routeGroup = nil

	return nil
}

// FindRouter 查询路由
func (s *Servant) FindRouter(message proto.Message) *router.Router {
	if s.routeGroup == nil {
		return nil
	}

	return s.routeGroup.Get(string(protoreflect.FullName(proto.MessageName(message))))
}

// ReplyMessage 回复消息给连接着
/*func (s *Servant) ReplyMessage(addr string, msg proto.Message) error {
	b, err := messages.MarshalMessageProtobuf(0, msg)
	if err != nil {
		panic(err)
	}

	bucket := s.vaddrs.getBucket(addr)
	clientId, ok := bucket.Get(addr)
	if !ok {
		goto repeat_label
	}

	if err := clientId.(*network.ClientID).PostUserMessage(b); err != nil {
		goto repeat_label
	}

	return nil
repeat_label:
	var resultErr error = nil
	repeat.Repeat(repeat.FnWithCounter(func(n int) error {
		if n >= 2 {
			return nil
		}

		bucket = s.vaddrs.getBucket(addr)
		clientId, ok = bucket.Get(addr)
		if !ok {
			resultErr = fmt.Errorf("post addr %s %s message unfound target", addr, reflect.TypeOf(msg))
			return resultErr
		}

		if err = clientId.(*network.ClientID).PostUserMessage(b); err != nil {
			resultErr = err
			return resultErr
		}

		resultErr = nil
		return nil
	}),
		repeat.StopOnSuccess(),
		repeat.WithDelay(repeat.ExponentialBackoff(500*time.Millisecond).Set()))
	return resultErr
}*/

/*func (s *Servant) onProxyConnected(conn *proxy.RpcProxyConn) {
	_, err := conn.RequestMessage(&internal.RegisterRequest{
		Vaddr: s.Config.Name + "@" + s.Config.VAddr,
	}, 2000)

	if err != nil {
		s.System.Error("proxy %s connected fail[error:%s]", conn.ToAddress(), err.Error())
		conn.Close()
		return
	}
}*/

func (s *Servant) spawConn(system *network.NetworkSystem) network.Client {
	return &ServantClientConn{
		Servant: s,
		recvice: circbuf.New(32768, &syncx.NoMutex{}),
	}
}
