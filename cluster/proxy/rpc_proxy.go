package proxy

import (
	"sync"

	"github.com/yamakiller/velcro-go/cluster/balancer"
	"github.com/yamakiller/velcro-go/rpc/rpcclient"
	"github.com/yamakiller/velcro-go/utils/host"
)

// NewRpcProxy 创建Rpc代理
// targetHosts 目标主机地址
// algorithm   均衡器算法
func NewRpcProxy(targetHosts []string, algorithm string) (*RpcProxy, error) {

	hosts := make([]string, 0)
	hostMap := make(map[string]*rpcclient.Conn)
	alive := make(map[string]bool)
	for _, targetHost := range targetHosts {
		_, _, err := host.Parse(targetHost)
		if err != nil {
			return nil, err
		}
	}

	lb, err := balancer.Build(algorithm, hosts)
	if err != nil {
		return nil, err
	}

	return &RpcProxy{
		hostMap:  hostMap,
		balancer: lb,
		alive:    alive,
	}, nil
}

type RpcProxy struct {
	// 连接器
	hostMap map[string]*rpcclient.Conn
	// 均衡器
	balancer balancer.Balancer

	// 活着的目标
	sync.RWMutex
	alive map[string]bool
}
