package proxy

import (
	"sync"
	"time"

	"github.com/yamakiller/velcro-go/cluster/balancer"
	"github.com/yamakiller/velcro-go/cluster/repeat"
	"github.com/yamakiller/velcro-go/logs"
	"github.com/yamakiller/velcro-go/rpc/client/asyn"
	"github.com/yamakiller/velcro-go/rpc/errs"
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

		//创建连接
		conn := &RpcProxyConn{}
		conn.Conn = asyn.NewConn(
			asyn.WithKleepalive(option.Kleepalive),
			asyn.WithConnected(conn.Connected),
			asyn.WithClosed(conn.Closed))

		alive[targetHost.VAddr] = false
		hostMap[targetHost.VAddr] = conn
		hostAddrMap[targetHost.VAddr] = targetHost.LAddr
	}

	lb, err := balancer.Build(option.Algorithm, hosts)
	if err != nil {
		return nil, err
	}

	return &RpcProxy{
		dialTimeouot: time.Millisecond * time.Duration(option.DialTimeout),
		hostMap:      hostMap,
		hostAddrMap:  hostAddrMap,
		balancer:     lb,
		alive:        alive,
		logger:       option.Logger,
		stopper:      make(chan struct{}),
	}, nil
}

// RpcProxy rpc 代理
type RpcProxy struct {
	// 连接等待超时
	dialTimeouot time.Duration
	// 连接器
	hostMap     map[string]*RpcProxyConn
	hostAddrMap map[string]string
	// 均衡器
	balancer balancer.Balancer
	// 代理连接成功回调函数
	connected func(*RpcProxyConn)
	// 活着的目标
	sync.RWMutex
	alive map[string]bool
	// 日志代理
	logger  logs.LogAgent
	stopper chan struct{}
}

// Shutdown 打开代理
func (rpx *RpcProxy) Open() {
	for host, conn := range rpx.hostMap {
		conn.proxy = rpx
		conn.repe = &RpcProxyConnRepeat{
			host:         rpx.hostAddrMap[host],
			conn:         conn,
			dialTimeouot: rpx.dialTimeouot,
			printError:   rpx.LogError,
		}

		rpx.LogInfo("%s connecting", rpx.hostAddrMap[host])
		if err := conn.Dial(rpx.hostAddrMap[host], rpx.dialTimeouot); err != nil {
			rpx.LogError("%s connect fail[error:%s]", rpx.hostAddrMap[host], err.Error())
			// 启动退避启动器
			conn.repe.start()
			continue
		}

		rpx.LogInfo("%s connected", rpx.hostAddrMap[host])

		rpx.alive[host] = true
		rpx.balancer.Add(host)
	}
}

// RequestMessage 集群请求消息
func (rpx *RpcProxy) RequestMessage(message proto.Message, timeout int64) (proto.Message, error) {
	var (
		host   string
		err    error
		future *asyn.Future
	)

	startMills := time.Now().UnixMilli()
	endMills := int64(0)
	host, err = rpx.balancer.Balance("")
	if err != nil {
		goto try_again_label
	}

	rpx.balancer.Inc(host)
	defer rpx.balancer.Done(host)

	future, err = rpx.hostMap[host].RequestMessage(message, timeout)
	if err != nil {
		goto try_again_label
	}

	future.Wait()
	if future.Error() != nil {
		if future.Error() == errs.ErrorRequestTimeout {
			return nil, errs.ErrorRequestTimeout
		}

		endMills = time.Now().UnixMilli()
		if endMills-startMills > timeout {
			return nil, errs.ErrorRequestTimeout
		}
		goto try_again_label
	}

	return future.Result(), nil
try_again_label:
	var result proto.Message = nil
	var resultErr error = nil
	repeat.Repeat(repeat.FnWithCounter(func(n int) error {
		if n >= 2 {
			return nil
		}

		endMills = time.Now().UnixMilli()
		if endMills-startMills > timeout {
			resultErr = errs.ErrorRequestTimeout
			return nil
		}

		host, err = rpx.balancer.Balance("")
		if err != nil {
			resultErr = err
			return resultErr
		}

		rpx.balancer.Inc(host)
		defer rpx.balancer.Done(host)

		future, err = rpx.hostMap[host].RequestMessage(message, timeout)
		if err != nil {
			resultErr = err
			return resultErr
		}

		future.Wait()
		if future.Error() != nil {
			if future.Error() == errs.ErrorRequestTimeout {
				resultErr = errs.ErrorRequestTimeout
				return nil
			}

			resultErr = err
			return resultErr
		}
		result = future.Result()
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
		conn.Destory()
	}
	rpx.hostMap = make(map[string]*RpcProxyConn)
	rpx.alive = make(map[string]bool)
	rpx.balancer = nil
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

func (rpx *RpcProxy) isStopped() bool {
	select {
	case <-rpx.stopper:
		return true
	default:
		return false
	}
}
