package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// Selector 顺序执行节点,只要有一个子节点
// 返回success或running,那么当前节点也返回success或running.
type Selector struct {
	core.Composite
}

func (stor *Selector) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("runningChild", 0, tick.GetTree().GetID(), stor.GetID())
}

func (stor *Selector) OnTick(tick *core.Tick) behavior.Status {

	var child = tick.Blackboard.GetInt("runningChild", tick.GetTree().GetID(), stor.GetID())
	for i := child; i < stor.GetChildCount(); i++ {
		var status = stor.GetChild(i).Execute(tick)
		if status != behavior.SUCCESS {
			if status == behavior.RUNNING {
				tick.Blackboard.Set("runningChild", i, tick.GetTree().GetID(), stor.GetID())
			}
			return status
		}
	}
	return behavior.SUCCESS

}
