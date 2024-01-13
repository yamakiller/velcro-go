package material

import "github.com/yamakiller/velcro-go/cluster/proxy"

type Router struct {
	VAddress string `yaml:"virtual-address"`

	MaxConn            int32                  `yaml:"max-conn"`                     // 最大连接数
	MaxIdleConn        int32                  `yaml:"max-idle-conn"`                // 最大空闲连接数
	MaxIdleConnTimeout int                    `yaml:"max-idle-conn-timeout-second"` // 空闲连接最大闲置时间(单位：秒)
	Rules              []int32                `yaml:"rules"`
	Commands           []string               `yaml:"commands"`
	Endpoints          []proxy.ResolveAddress `yaml:"endpoints"`
}
