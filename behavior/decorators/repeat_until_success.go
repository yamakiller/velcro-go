package decorators

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// 重复N此执行节点直到遇到非FAILURE返回，或者达到尝试次数上限.
type RepeatUntilSuccess struct {
	core.Decorator
	maxLoop int
}

// Initialize 初始化
// Settings 属性:
//
//	      key:   maxLoop
//		  value: 尝试次数
func (rpu *RepeatUntilSuccess) Initialize() {
	rpu.Decorator.Initialize()
	//this.maxLoop = setting.GetPropertyAsInt("maxLoop")
	//if this.maxLoop < 1 {
	//	panic("maxLoop parameter in MaxTime decorator is an obligatory parameter")
	//}
}

func (rpu *RepeatUntilSuccess) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("i", 0, tick.GetTree().GetID(), rpu.GetID())
}

func (rpu *RepeatUntilSuccess) OnTick(tick *core.Tick) behavior.Status {
	if rpu.GetChild() == nil {
		return behavior.ERROR
	}
	var i = tick.Blackboard.GetInt("i", tick.GetTree().GetID(), rpu.GetID())
	var status = behavior.ERROR
	for rpu.maxLoop < 0 || i < rpu.maxLoop {
		status = rpu.GetChild().Execute(tick)
		if status == behavior.FAILURE {
			i++
		} else {
			break
		}
	}

	tick.Blackboard.Set("i", i, tick.GetTree().GetID(), rpu.GetID())
	return status
}
