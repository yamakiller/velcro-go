package gateway

import (
	"errors"
	"reflect"
	"sync"

	"github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/cluster/proxy"
	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/network"
	"google.golang.org/protobuf/proto"
)

func New(options ...GatewayConfigOption) *Gateway {

	config := Configure(options...)

	g := &Gateway{clients: make(map[network.CIDKEY]Client), logger: config.Logger}
	g.System = config.NewNetworkSystem(
		network.WithMetricProviders(config.MetricsProvider),
		network.WithLoggerFactory(func(System *network.NetworkSystem) logs.LogAgent {
			return g.logger
		}),
		network.WithNetworkTimeout(config.NetowkTimeout),
		network.WithProducer(g.newClient),
		network.WithVAddr(config.VAddr),
	)

	if config.ClientPool == nil {
		config.ClientPool = NewDefaultGatewayClientPool(g, config.MessageMaxTimeout)
	}

	g.clientPool = config.ClientPool
	g.onlineOfNumber = config.OnlineOfNumber
	g.routeURI = config.RouterURI
	g.routeProxyFrequency = config.RouteProxyFrequency
	g.routeProxyDialTimeout = config.RouteProxyDialTimeout
	g.routeProxyKleepalive = config.RouteProxyKleepalive
	g.routeProxyAlgorithm = config.RouteProxyAlgorithm
	g.vaddr = config.VAddr
	g.laddr = config.LAddr
	return g
}

// Gateway 网关
type Gateway struct {
	vaddr  string // 网关局域网虚拟地址
	laddr  string // 网关监听地址
	System *network.NetworkSystem
	// 连接客户端
	clients        map[network.CIDKEY]Client
	onlineOfNumber int
	mu             sync.Mutex
	clientPool     GatewayClientPool
	// 编码
	encryption *Encryption
	// 路由
	routeURI              string
	routeProxyFrequency   int32
	routeProxyDialTimeout int32
	routeProxyKleepalive  int32
	routeProxyAlgorithm   string
	routeGroup            *router.RouterGroup
	logger                logs.LogAgent
}

func (g *Gateway) Start() error {
	r, err := router.Loader(g.routeURI,
		router.WithFrequency(g.routeProxyFrequency),
		router.WithAlgorithm(g.routeProxyAlgorithm),
		router.WithDialTimeout(g.routeProxyDialTimeout),
		router.WithKleepalive(g.routeProxyKleepalive),
		router.WithConnectedCallback(g.onProxyConnected),
		router.WithReceiveCallback(g.onProxyRecvice),
	)

	if err != nil {
		return err
	}

	// 尝试连接服务
	g.routeGroup = r
	g.routeGroup.Open()
	// 打开监听
	if err := g.System.Open(g.laddr); err != nil {
		return err
	}

	return nil
}

func (g *Gateway) Stop() error {

	if g.System != nil {
		g.System.Info("Gateway Shutdown Closing client")

		g.mu.Lock()
		for _, c := range g.clients {
			c.ClientID().UserClose()
		}
		g.mu.Unlock()
		g.System.Info("Gateway Shutdown Closed client")
		g.System.Shutdown()
		g.System.Info("Gateway Shutdown")
		g.System = nil
	}

	if g.routeGroup != nil {
		g.routeGroup.Shutdown()
	}
	return nil
}

func (g *Gateway) Register(cid *network.ClientID, l Client) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.clients) >= g.onlineOfNumber {
		return errors.New("gateway: client fulled")
	}

	_, ok := g.clients[network.Key(cid)]
	if ok {
		panic("register linker error: ClientID Repeat")
	}

	l.referenceIncrement()
	g.clients[network.Key(cid)] = l

	return nil
}

func (g *Gateway) UnRegister(cid *network.ClientID) {

	g.mu.Lock()
	l, ok := g.clients[network.Key(cid)]
	if !ok {
		g.mu.Unlock()
		return
	}

	ref := l.referenceDecrement()
	delete(g.clients, network.Key(cid))

	g.mu.Unlock()

	if ref == 0 {
		g.free(l)
	}
}

func (g *Gateway) GetClient(cid *network.ClientID) Client {
	g.mu.Lock()
	defer g.mu.Unlock()

	l, ok := g.clients[network.Key(cid)]
	if !ok {
		return nil
	}

	l.referenceIncrement()
	return l
}

func (g *Gateway) ReleaseClient(c Client) {
	g.mu.Lock()
	ref := c.referenceDecrement()
	g.mu.Unlock()

	if ref == 0 {
		g.free(c)
	}
}

// FindRouter 查询路由
func (g *Gateway) FindRouter(message proto.Message) *router.Router {
	return g.routeGroup.Get(string(proto.MessageName(message)))
}

// NewClientID 创建客户端连接者ID
func (g *Gateway) NewClientID(id string) *network.ClientID {
	return g.System.NewClientID(id)
}

func (g *Gateway) onProxyConnected(conn *proxy.RpcProxyConn) {
	_, err := conn.RequestMessage(&protocols.RegisterRequest{
		Vaddr: g.vaddr,
	}, 2000)
	if err != nil {
		g.System.Error("proxy %s connected fail[error:%s]", conn.ToAddress(), err.Error())
		conn.Close()
		return
	}
}

func (g *Gateway) onProxyRecvice(message interface{}) {
	switch msg := message.(type) {
	case *protocols.Backward:
		g.onBackwardClient(msg)
	case *protocols.UpdateRule:
		g.onUpdateRuleClient(msg)
	case *protocols.Closing:
		g.onCloseClient(msg)
	default:
		g.System.Error("Unknown %s message received from service", reflect.TypeOf(msg).Name())
	}
}

func (g *Gateway) onBackwardClient(backward *protocols.Backward) {
	if backward.Target == nil {
		return
	}

	c := g.GetClient(backward.Target)
	if c == nil {
		return
	}

	defer g.ReleaseClient(c)
	anyMsg := backward.Msg.ProtoReflect().New().Interface()

	if err := backward.Msg.UnmarshalTo(anyMsg); err != nil {
		g.System.Error("forward %s message unmarshal fail[error: %s]",
			string(backward.Msg.MessageName()),
			err.Error())
		return
	}

	c.Post(anyMsg)
}

func (g *Gateway) onUpdateRuleClient(updateRule *protocols.UpdateRule) {
	c := g.GetClient(updateRule.Target)
	if c == nil {
		g.System.Debug("update rule fail unfound client id[%+v]", updateRule.Target)
		return
	}
	defer g.ReleaseClient(c)
	c.onUpdateRule(updateRule.Rule)
}

func (g *Gateway) onCloseClient(closing *protocols.Closing) {
	c := g.GetClient(closing.ClientID)
	if c == nil {
		return
	}
	defer g.ReleaseClient(c)

	c.ClientID().UserClose()
	// 广播所有服务器关闭套接字
	g.routeGroup.Broadcast(&protocols.Closed{ClientID: closing.ClientID})
}

func (g *Gateway) newClient(System *network.NetworkSystem) network.Client {
	return g.clientPool.Get()
}

func (g *Gateway) free(l Client) {
	l.Destory()
	g.clientPool.Put(l)
}
