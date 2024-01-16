package material

import "github.com/yamakiller/velcro-go/cluster/proxy"

type Router struct {
	VAddress           string                 `yaml:"virtual-address"`              // 目标虚拟标记
	MaxConn            int32                  `yaml:"max-conn"`                     // 最大连接数
	MaxIdleConn        int32                  `yaml:"max-idle-conn"`                // 最大空闲连接数
	MaxIdleConnTimeout int                    `yaml:"max-idle-conn-timeout-minute"` // 空闲连接最大闲置时间(单位：分钟)
	Rules              []int32                `yaml:"rules"`                        //用户通过权限
	Endpoints          []proxy.ResolveAddress `yaml:"endpoints"`                    //目标连接地址和虚拟地址
	Commands           []string               `yaml:"commands"`                     //可通行命令
}
