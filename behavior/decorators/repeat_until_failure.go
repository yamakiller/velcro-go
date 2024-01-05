package decorators

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// 重复N此执行节点直到遇到非SUCCESS返回，或者达到尝试次数上限.
type RepeatUntilFailure struct {
	core.Decorator
	maxLoop int
}

func (rpu *RepeatUntilFailure) Initialize() {
	rpu.Decorator.Initialize()
	//rpu.maxLoop = setting.GetPropertyAsInt("maxLoop")
	//if this.maxLoop < 1 {
	//	panic("maxLoop parameter in MaxTime decorator is an obligatory parameter")
	//}
}

func (rpu *RepeatUntilFailure) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("i", 0, tick.GetTree().GetID(), rpu.GetID())
}

func (rpu *RepeatUntilFailure) OnTick(tick *core.Tick) behavior.Status {
	if rpu.GetChild() == nil {
		return behavior.ERROR
	}
	var i = tick.Blackboard.GetInt("i", tick.GetTree().GetID(), rpu.GetID())
	var status = behavior.ERROR
	for rpu.maxLoop < 0 || i < rpu.maxLoop {
		status = rpu.GetChild().Execute(tick)
		if status == behavior.SUCCESS {
			i++
		} else {
			break
		}
	}

	tick.Blackboard.Set("i", i, tick.GetTree().GetID(), rpu.GetID())
	return status
}
