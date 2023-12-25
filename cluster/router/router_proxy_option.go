package router

import (
	"github.com/yamakiller/velcro-go/cluster/proxy"
	"github.com/yamakiller/velcro-go/logs"
)

// ConnConfigOption 是一个配置rpc connector 的函数
type RouterRpcProxyConfigOption func(option *RouterRpcProxyOption)

func Configure(options ...RouterRpcProxyConfigOption) *RouterRpcProxyOption {
	config := defaultRouterRpcProxyOption()
	for _, option := range options {
		option(config)
	}

	return config
}

// WithKleepalive 设置连接器保活时间(单位:毫秒)
func WithKleepalive(kleepalive int32) RouterRpcProxyConfigOption {

	return func(opt *RouterRpcProxyOption) {
		opt.Kleepalive = kleepalive
	}
}

// WithDialTimeout 设置连接器连接等待超时时间(单位:毫秒)
func WithDialTimeout(timeout int32) RouterRpcProxyConfigOption {
	return func(opt *RouterRpcProxyOption) {
		opt.DialTimeout = timeout
	}
}

// WithLogger 设置日志代理
func WithLogger(logger logs.LogAgent) RouterRpcProxyConfigOption {
	return func(opt *RouterRpcProxyOption) {
		opt.Logger = logger
	}
}

// WithConnectedCallback 设置连接成功后的回调函数
func WithConnectedCallback(f func(*proxy.RpcProxyConn)) RouterRpcProxyConfigOption {
	return func(opt *RouterRpcProxyOption) {
		opt.ConnectedCallback = f
	}
}

// WithReceiveCallback 设置连接器消息回调函数
func WithReceiveCallback(f func(interface{})) RouterRpcProxyConfigOption {
	return func(opt *RouterRpcProxyOption) {
		opt.RecviceCallback = f
	}
}

// WithAlgorithm 设置平衡器算法:ip-hash、consistent-hash、p2c、
// random、round-robin、least-load、bounded.
func WithAlgorithm(algorithm string) RouterRpcProxyConfigOption {
	return func(opt *RouterRpcProxyOption) {
		opt.Algorithm = algorithm
	}
}

// RouterRpcProxyOption 路由rpc proxy 参数设置
type RouterRpcProxyOption struct {
	Kleepalive        int32         // 连接器保活时间(单位:毫秒)
	DialTimeout       int32         // 连接器连接等待超时时间(单位:毫秒)
	Logger            logs.LogAgent // 日志代理
	ConnectedCallback func(*proxy.RpcProxyConn)
	RecviceCallback   func(interface{})
	Algorithm         string
}

func defaultRouterRpcProxyOption() *RouterRpcProxyOption {
	return &RouterRpcProxyOption{
		Kleepalive:  10 * 1000,
		DialTimeout: 2 * 1000,
		Algorithm:   "p2c",
	}
}
