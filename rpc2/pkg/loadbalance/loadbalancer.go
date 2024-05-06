package loadbalance

import (
	"context"

	"github.com/yamakiller/velcro-go/rpc2/pkg/discovery"
)

// Picker 为下一个 RPC 调用选择一个实例.
type Picker interface {
	Next(ctx context.Context, request interface{}) discovery.Instance
}

// Loadbalancer 为给定的服务发现结果生成选择器.
type Loadbalancer interface {
	GetPicker(discovery.Result) Picker
	Name() string // unique key
}

// Rebalancer 是一种负载均衡器，当服务发现结果发生变化时执行重新均衡.
type Rebalancer interface {
	Rebalance(discovery.Change)
	Delete(discovery.Change)
}
