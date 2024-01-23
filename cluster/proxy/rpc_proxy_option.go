package proxy

// ConnConfigOption 是一个配置rpc connector 的函数
type RpcProxyConfigOption func(option *RpcProxyOption)

func Configure(options ...RpcProxyConfigOption) *RpcProxyOption {
	config := defaultRpcProxyOption()
	for _, option := range options {
		option(config)
	}

	return config
}

// WithPoolConfig 连接池配置
/*func WithPoolConfig(cfg client.LongConnPoolConfig) RpcProxyConfigOption {
	return func(opt *RpcProxyOption) {
		opt.PoolConfig = cfg
	}
}*/

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
	//PoolConfig client.LongConnPoolConfig
	TargetHost []ResolveAddress
	Algorithm  string
}

type ResolveAddress struct {
	VAddr string `yaml:"vaddr"` // 虚拟地址
	LAddr string `yaml:"laddr"` // 真实地址
}

func defaultRpcProxyOption() *RpcProxyOption {
	return &RpcProxyOption{
		//PoolConfig: client.LongConnPoolConfig{},
		Algorithm: "p2c",
	}
}
