package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// Sequenece 顺序执行节点,只要有一个子节点返回failed或running,当前节点也返回failed或running.
type Sequenece struct {
	core.Composite
}

func (sn *Sequenece) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("runningChild", 0, tick.GetTree().GetID(), sn.GetID())
}

func (ms *Sequenece) OnTick(tick *core.Tick) behavior.Status {
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
