package route

import (
	"sync"

	cmap "github.com/orcaman/concurrent-map"
)

type Balancer struct {
	Label           string
	VxAddrss        string
	Rules           cmap.ConcurrentMap // 角色表
	Command         cmap.ConcurrentMap // 指令表
	activedEndpoint int                // 当前活动的节点
	endpoints       []Endpoint         // 健康的活动节点
	targets         []string           // 配置的目标节点
	mu              sync.RWMutex       // endpoints 节点读写锁
}
