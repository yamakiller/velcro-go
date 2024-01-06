package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// SerialSelector 每次都按顺序执行所有节点, 只要有一个子
// 节点返回success, 那么当前节点也返回success, 否则返回failure.
type SerialSelector struct {
	core.Composite
}

func (sstor *SerialSelector) OnTick(tick *core.Tick) behavior.Status {
	for i := 0; i < sstor.GetChildCount(); i++ {
		var status = sstor.GetChild(i).Execute(tick)

		if status == behavior.SUCCESS {
			return status
		}
	}
	return behavior.FAILURE
}
