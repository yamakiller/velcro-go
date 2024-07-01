package proxy

import (
	"fmt"
	"sync"
	"time"

	"github.com/yamakiller/velcro-go/cluster/balancer"
	"github.com/yamakiller/velcro-go/cluster/repeat"
	"github.com/yamakiller/velcro-go/rpc/client"
	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
)

// NewRpcProxy 创建Rpc代理
func NewRpcProxy(options ...RpcProxyConfigOption) (*RpcProxy, error) {
	opt := Configure(options...)
	return NewRpcProxyOption(opt)
}

func NewRpcProxyOption(option *RpcProxyOption) (*RpcProxy, error) {
	hosts := make([]string, 0)
	hostMap := make(map[string]*RpcProxyConn)
	hostAddrMap := make(map[string]string)
	alive := make(map[string]bool)

	for _, targetHost := range option.TargetHost {
		conn := &RpcProxyConn{}

		alive[targetHost.VAddr] = false
		hostMap[targetHost.VAddr] = conn
		hostAddrMap[targetHost.VAddr] = targetHost.LAddr
	}

	lb, err := balancer.Build(option.Algorithm, hosts)
	if err != nil {
		return nil, err
	}

	return &RpcProxy{
		poolConfig:  &option.PoolConfig,
		hostMap:     hostMap,
		hostAddrMap: hostAddrMap,
		balancer:    lb,
		alive:       alive,
		stopper:     make(chan struct{}),
	}, nil
}

// RpcProxy rpc 代理
type RpcProxy struct {
	// 连接等待超时
	poolConfig *client.LongConnPoolConfig
	// 连接器
	hostMap     map[string]*RpcProxyConn
	hostAddrMap map[string]string
	// 均衡器
	balancer balancer.Balancer
	// 活着的目标
	sync.RWMutex
	alive   map[string]bool
	stopper chan struct{}
}

// Shutdown 打开代理
func (rpx *RpcProxy) Open() {
	for host, conn := range rpx.hostMap {
		conn.proxy = rpx
		conn.LongConnPool = client.NewDefaultConnPool(rpx.hostAddrMap[host], client.LongConnPoolConfig{
			MaxConn: rpx.poolConfig.MaxConn,
			MaxIdleConn: rpx.poolConfig.MaxIdleConn,
			MaxIdleConnTimeout: rpx.poolConfig.MaxIdleConnTimeout,
		})

		vlog.Infof("%s connecting", rpx.hostAddrMap[host])
		vlog.Infof("%s connected", rpx.hostAddrMap[host])

		rpx.alive[host] = true
		rpx.balancer.Add(host)
	}
}

// RequestMessage 集群请求消息
func (rpx *RpcProxy) RequestMessage(message proto.Message, timeout int64) (proto.Message, error) {
	if message ==nil{
		panic(fmt.Errorf("(rpx *RpcProxy) RequestMessage  message  is nil"))
	}
	var (
		host string
		result    proto.Message = nil
		resultErr error = nil
	)

	startMills := time.Now().UnixMilli()
	endMills := int64(0)
	host, resultErr = rpx.balancer.Balance("")
	if resultErr != nil {
		goto try_again_label
	}

	rpx.balancer.Inc(host)
	defer rpx.balancer.Done(host)

	result, resultErr = rpx.hostMap[host].RequestMessage(message, timeout)

	if resultErr != nil {
		if resultErr == errs.ErrorRequestTimeout {
			return nil, errs.ErrorRequestTimeout
		}

		endMills = time.Now().UnixMilli()
		if endMills-startMills > timeout {
			return nil, errs.ErrorRequestTimeout
		}
		goto try_again_label
	}

	return result, nil
try_again_label:

	repeat.Repeat(repeat.FnWithCounter(func(n int) error {
		if n >= 2 {
			return nil
		}

		endMills = time.Now().UnixMilli()
		if endMills-startMills > timeout {
			resultErr = errs.ErrorRequestTimeout
			return nil
		}

		host, resultErr = rpx.balancer.Balance("")
		if resultErr != nil {
			return resultErr
		}

		rpx.balancer.Inc(host)
		defer rpx.balancer.Done(host)

		result, resultErr = rpx.hostMap[host].RequestMessage(message, timeout)

		if resultErr != nil {
			if resultErr == errs.ErrorRequestTimeout {
				resultErr = errs.ErrorRequestTimeout
				return nil
			}

			return resultErr
		}
		return nil
	}),
		repeat.StopOnSuccess(),
		repeat.WithDelay(repeat.ExponentialBackoff(200*time.Millisecond).Set()))

	return result, resultErr
}

// Shutdown 关闭代理
func (rpx *RpcProxy) Shutdown() {
	//释放资源
	close(rpx.stopper)
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
