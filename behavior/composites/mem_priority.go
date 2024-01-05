package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// MemPriority 顺序执行，只有一个不是失败就返回
type MemPriority struct {
	core.Composite
}

func (mpri *MemPriority) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("runningChild", 0, tick.GetTree().GetID(), mpri.GetID())
}

func (mpri *MemPriority) OnTick(tick *core.Tick) behavior.Status {
	var child = tick.Blackboard.GetInt("runningChild", tick.GetTree().GetID(), mpri.GetID())
	for i := child; i < mpri.GetChildCount(); i++ {
		var status = mpri.GetChild(i).Execute(tick)

		if status != behavior.FAILURE {
			if status == behavior.RUNNING {
				tick.Blackboard.Set("runningChild", i, tick.GetTree().GetID(), mpri.GetID())
			}

			return status
		}
	}
	return behavior.FAILURE
}
