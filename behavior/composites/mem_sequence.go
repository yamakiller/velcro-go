package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// MemSequence 顺序执行，只要一个不成功就返回
type MemSequence struct {
	core.Composite
}

func (ms *MemSequence) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("runningChild", 0, tick.GetTree().GetID(), ms.GetID())
}

func (ms *MemSequence) OnTick(tick *core.Tick) behavior.Status {
	var child = tick.Blackboard.GetInt("runningChild", tick.GetTree().GetID(), ms.GetID())
	for i := child; i < ms.GetChildCount(); i++ {
		var status = ms.GetChild(i).Execute(tick)

		if status != behavior.SUCCESS {
			if status == behavior.RUNNING {
				tick.Blackboard.Set("runningChild", i, tick.GetTree().GetID(), ms.GetID())
			}

			return status
		}
	}
	return behavior.SUCCESS
}
