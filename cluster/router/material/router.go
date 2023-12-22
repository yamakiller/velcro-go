package material

import "github.com/yamakiller/velcro-go/cluster/proxy"

type Router struct {
	VAddress  string        `yaml:"virtual-address"`
	Decided   proxy.Decided `yaml:"decided"`
	Rules     []int32       `yaml:"rules"`
	Commands  []string      `yaml:"commands"`
	Endpoints []string      `yaml:"endpoints"`
}
