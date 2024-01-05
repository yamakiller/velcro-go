package material

import "github.com/yamakiller/velcro-go/cluster/proxy"

type Router struct {
	VAddress           string                 `yaml:"virtual-address"`
	MaxIdleGlobal      int32                  `yaml:"max-idle-conn"`                // 最大连接数
	MaxIdleTimeout     int                    `yaml:"max-idle-timeout-minute"`      // 连接最大空闲时间(单位:分钟)
	MaxIdleConnTimeout int                    `yaml:"max-idle-conn-timeout-second"` // 连接最大超时时间(单位:秒)
	Kleepalive         int32                  `yaml:"kleepalive-millisec"`          // 心跳时间(单位:毫秒)
	Rules              []int32                `yaml:"rules"`
	Commands           []string               `yaml:"commands"`
	Endpoints          []proxy.ResolveAddress `yaml:"endpoints"`
}
