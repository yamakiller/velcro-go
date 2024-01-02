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

// WithLogger 设置日志代理
func WithLogger(logger logs.LogAgent) RpcProxyConfigOption {
	return func(opt *RpcProxyOption) {
		opt.Logger = logger
	}
}

// WithTargetHost 设置目标主机组
func WithTargetHost(host []ResolveAddress) RpcProxyConfigOption {
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
	Kleepalive  int32         // 连接器保活时间(单位:毫秒)
	DialTimeout int32         // 连接器连接等待超时时间(单位:毫秒)
	Logger      logs.LogAgent // 日志代理
	TargetHost  []ResolveAddress
	Algorithm   string
}

type ResolveAddress struct {
	VAddr string `yaml:"vaddr"` // 虚拟地址
	LAddr string `yaml:"laddr"` // 真实地址
}

func defaultRpcProxyOption() *RpcProxyOption {
	return &RpcProxyOption{
		Kleepalive:  10 * 1000,
		DialTimeout: 2 * 1000,
		Algorithm:   "p2c",
	}
}
