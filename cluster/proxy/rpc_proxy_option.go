package proxy

import (
	"github.com/yamakiller/velcro-go/logs"
)

// ConnConfigOption 是一个配置rpc connector 的函数
type RpcProxyConfigOption func(option *RpcProxyOption)

func Configure(options ...RpcProxyConfigOption) *RpcProxyOption {
	config := defaultRpcProxyOption()
	for _, option := range options {
		option(config)
	}

	return config
}

// WithKleepalive 设置连接器保活时间(单位:毫秒)
func WithKleepalive(kleepalive int32) RpcProxyConfigOption {

	return func(opt *RpcProxyOption) {
		opt.Kleepalive = kleepalive
	}
}

// WithDialTimeout 设置连接器连接等待超时时间(单位:毫秒)
func WithDialTimeout(timeout int32) RpcProxyConfigOption {
	return func(opt *RpcProxyOption) {
		opt.DialTimeout = timeout
	}
}

// WithFrequency 设置连接器重连频率(时间-单位:毫秒)
func WithFrequency(frequency int32) RpcProxyConfigOption {
	return func(opt *RpcProxyOption) {
		opt.Frequency = frequency
	}
}

// WithLogger 设置日志代理
func WithLogger(logger logs.LogAgent) RpcProxyConfigOption {
	return func(opt *RpcProxyOption) {
		opt.Logger = logger
	}
}

func WithConnectedCallback(f func(*RpcProxyConn)) RpcProxyConfigOption {
	return func(opt *RpcProxyOption) {
		opt.ConnectedCallback = f
	}
}

// WithReceiveCallback 设置连接器消息回调函数
func WithReceiveCallback(f func(interface{})) RpcProxyConfigOption {
	return func(opt *RpcProxyOption) {
		opt.RecviceCallback = f
	}
}

// WithTargetHost 设置目标主机组
func WithTargetHost(host []string) RpcProxyConfigOption {
	return func(opt *RpcProxyOption) {
		opt.TargetHost = host
	}
}

// WithAlgorithm 设置平衡器算法:ip-hash、consistent-hash、p2c、
// random、round-robin、least-load、bounded.
func WithAlgorithm(algorithm string) RpcProxyConfigOption {
	return func(opt *RpcProxyOption) {
		opt.Algorithm = algorithm
	}
}

// RpcProxyOption
type RpcProxyOption struct {
	Kleepalive        int32         // 连接器保活时间(单位:毫秒)
	DialTimeout       int32         // 连接器连接等待超时时间(单位:毫秒)
	Frequency         int32         // 连接检查频率/自动重连频率(单位:毫秒)
	Logger            logs.LogAgent // 日志代理
	ConnectedCallback func(*RpcProxyConn)
	RecviceCallback   func(interface{})

	TargetHost []string
	Algorithm  string
}

func defaultRpcProxyOption() *RpcProxyOption {
	return &RpcProxyOption{
		Kleepalive:  10 * 1000,
		DialTimeout: 1 * 1000,
		Frequency:   2 * 1000,
		Algorithm:   "p2c",
	}
}
