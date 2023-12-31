package material

import "github.com/yamakiller/velcro-go/cluster/proxy"

type Router struct {
	VAddress  string                 `yaml:"virtual-address"`
	Rules     []int32                `yaml:"rules"`
	Commands  []string               `yaml:"commands"`
	Endpoints []proxy.ResolveAddress `yaml:"endpoints"`
}
