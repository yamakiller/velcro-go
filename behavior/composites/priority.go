package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// Priority 顺序执行，只要一个不是失败就返回
type Priority struct {
	core.Composite
}

func (pri *Priority) OnTick(tick *core.Tick) behavior.Status {
	for i := 0; i < pri.GetChildCount(); i++ {
		var status = pri.GetChild(i).Execute(tick)
		if status != behavior.FAILURE {
			return status
		}
	}
	return behavior.FAILURE
}
