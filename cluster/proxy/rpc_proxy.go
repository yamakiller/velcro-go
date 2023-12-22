package proxy

import (
	"sync"
	"time"

	"github.com/yamakiller/velcro-go/cluster/balancer"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/rpc/rpcclient"
	"github.com/yamakiller/velcro-go/utils/host"
)

// NewRpcProxy 创建Rpc代理
func NewRpcProxy(options ...RpcProxyConfigOption) (*RpcProxy, error) {
	opt := Configure(options...)

	return NewRpcProxyOption(opt)
}

func NewRpcProxyOption(option *RpcProxyOption) (*RpcProxy, error) {
	hosts := make([]string, 0)
	hostMap := make(map[string]*RpcProxyConn)
	alive := make(map[string]bool)

	for _, targetHost := range option.TargetHost {
		_, _, err := host.Parse(targetHost)
		if err != nil {
			return nil, err
		}

		//创建连接
		conn := &RpcProxyConn{}
		conn.Conn = rpcclient.NewConn(
			rpcclient.WithKleepalive(option.Kleepalive),
			rpcclient.WithConnected(conn.Connected),
			rpcclient.WithClosed(conn.Closed),
			rpcclient.WithReceive(conn.Receive))

		// 初始化时, 未处于活动状态
		alive[targetHost] = false
		hostMap[targetHost] = conn
	}

	lb, err := balancer.Build(option.Algorithm, hosts)
	if err != nil {
		return nil, err
	}

	return &RpcProxy{
		frequency:    time.Millisecond * time.Duration(option.Frequency),
		dialTimeouot: time.Millisecond * time.Duration(option.DialTimeout),
		hostMap:      hostMap,
		balancer:     lb,
		connected:    option.ConnectedCallback,
		recvice:      option.RecviceCallback,
		alive:        alive,
		logger:       option.Logger,
		stopper:      make(chan struct{}),
	}, nil
}

// RpcProxy rpc 代理
type RpcProxy struct {
	// 重连尝试频率
	frequency time.Duration
	// 连接等待超时
	dialTimeouot time.Duration
	// 连接器
	hostMap map[string]*RpcProxyConn
	// 均衡器
	balancer balancer.Balancer
	// 代理连接成功回调函数
	connected func(*RpcProxyConn)
	// 代理接收数据回调函数
	recvice func(interface{})
	// 活着的目标
	sync.RWMutex
	alive map[string]bool
	// 日志代理
	logger logs.LogAgent
	// 停止器
	stopper chan struct{}
	done    sync.WaitGroup
}

// Shutdown 打开代理
func (rpx *RpcProxy) Open() {
	for host, conn := range rpx.hostMap {
		conn.proxy = rpx
		rpx.LogInfo("%s connecting", host)
		if err := conn.Dial(host, rpx.dialTimeouot); err != nil {
			rpx.LogError("%s connect fail[error:%s]", host, err.Error())
			continue
		}

		rpx.LogInfo("%s connected", host)

		rpx.alive[host] = true
		rpx.balancer.Add(host)
	}

	go rpx.guardian()
}

// RequestMessage 集群请求消息
func (rpx *RpcProxy) RequestMessage(message interface{}, timeout int64) (interface{}, error) {
	host, err := rpx.balancer.Balance("")
	if err != nil {
		rpx.LogError("RequestMessage %v fail[error:%s]", message, err.Error())
		return nil, err
	}

	rpx.balancer.Inc(host)
	defer rpx.balancer.Done(host)
	return rpx.hostMap[host].RequestMessage(message, uint64(timeout))
}

// PostMessage 集群推送消息
func (rpx *RpcProxy) PostMessage(message interface{}) error {
	host, err := rpx.balancer.Balance("")
	if err != nil {

		return err
	}
	rpx.balancer.Inc(host)
	defer rpx.balancer.Done(host)
	return rpx.hostMap[host].PostMessage(message)
}

// Shutdown 关闭代理
func (rpx *RpcProxy) Shutdown() {
	close(rpx.stopper)
	rpx.done.Wait()

	//释放资源
	for _, conn := range rpx.hostMap {
		conn.Close()
		conn.Destory()
	}
	rpx.hostMap = make(map[string]*RpcProxyConn)
	rpx.alive = make(map[string]bool)
	rpx.balancer = nil
}

func (rpx *RpcProxy) isStopped() bool {
	select {
	case <-rpx.stopper:
		return true
	default:
		return false
	}
}

func (rpx *RpcProxy) guardian() {
	defer rpx.done.Done()
	for {
		if rpx.isStopped() {
			break
		}

		rpx.RLock()
		for host, isAlive := range rpx.alive {
			if isAlive {
				// rpx.RUnlock()
				continue
			}
			if _,ok := rpx.hostMap[host];!ok {
				// rpx.RUnlock()
				continue
			}
			rpx.LogInfo("%s reconnecting", host)
			if err := rpx.hostMap[host].Dial(host, rpx.dialTimeouot); err != nil {
				// rpx.RUnlock()
				rpx.LogInfo("%s reconnect fail[error:%s]", host, err.Error())
				continue
			}
			rpx.RUnlock()
			rpx.LogInfo("%s reconnected", host)

			rpx.Lock()
			rpx.alive[host] = true
			rpx.Unlock()

		
			rpx.balancer.Add(host)

			rpx.RLock()
		}
		rpx.RUnlock()
		time.Sleep(rpx.frequency)
	}
}

func (rpx *RpcProxy) LogError(fmts string, args ...interface{}) {
	if rpx.logger == nil {
		return
	}
	rpx.logger.Error("[RPCPROXY]", fmts, args...)
}

func (rpx *RpcProxy) LogInfo(fmts string, args ...interface{}) {
	if rpx.logger == nil {
		return
	}
	rpx.logger.Info("[RPCPROXY]", fmts, args...)
}

// logdebug 内部调用debug日志
func (rpx *RpcProxy) LogDebug(fmts string, args ...interface{}) {
	if rpx.logger == nil {
		return
	}
	rpx.logger.Debug("[RPCPROXY]", fmts, args...)
}
