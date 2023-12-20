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
	config := Configure(options...)

	return NewRpcProxyOption(config)
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

		//1.创建连接
		conn := &RpcProxyConn{}
		conn.Conn = rpcclient.NewConn(
			rpcclient.WithKleepalive(option.Kleepalive),
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
		recvice:      option.RecviceCallback,
		alive:        alive,
		logger:       option.Logger,
		stopper:      make(chan struct{}),
	}, nil
}

type RpcProxy struct {
	// 重连尝试频率
	frequency time.Duration
	// 连接等待超时
	dialTimeouot time.Duration
	// 连接器
	hostMap map[string]*RpcProxyConn
	// 均衡器
	balancer balancer.Balancer
	// 代理接收数据回调函数
	recvice func(interface{})
	// 活着的目标
	sync.RWMutex
	alive map[string]bool

	//
	logger logs.LogAgent

	stopper chan struct{}
	done    sync.WaitGroup
}

func (rpx *RpcProxy) Open() {
	for host, conn := range rpx.hostMap {
		conn.proxy = rpx
		if err := conn.Dial(host, rpx.dialTimeouot); err != nil {
			if rpx.logger != nil {
				rpx.logger.Error("[RPCPROXY]", "%s connect fail[error:%s]", host, err.Error())
			}
			continue
		}

		rpx.alive[host] = true
		rpx.balancer.Add(host)
	}

	go rpx.guardian()
}

func (rpx *RpcProxy) RequestMessage(message interface{}, timeout int64) (interface{}, error) {
	host, err := rpx.balancer.Balance("")
	if err != nil {
		rpx.logger.Error("[RPCPROXY]", "RequestMessage %v fail[error:%s]", message, err.Error())
		return nil, err
	}

	rpx.balancer.Inc(host)
	defer rpx.balancer.Done(host)
	return rpx.hostMap[host].RequestMessage(message, uint64(timeout))
}

func (rpx *RpcProxy) Shutdown() {
	close(rpx.stopper)
	rpx.done.Wait()

	//释放资源
	for _, conn := range rpx.hostMap {
		conn.Close()
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
				rpx.RUnlock()
				continue
			}

			if err := rpx.hostMap[host].Dial(host, rpx.dialTimeouot); err != nil {
				rpx.RUnlock()
				continue
			}

			rpx.RUnlock()

			rpx.Lock()
			rpx.alive[host] = true
			rpx.Unlock()

			rpx.balancer.Add(host)
		}

		time.Sleep(rpx.frequency)
	}
}
