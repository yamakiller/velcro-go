package decorators

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

/**
 * This decorator limit the number of times its child can be called. After a
 * certain number of times, the Limiter decorator returns `FAILURE` without
 * executing the child.
 *
 * @module b3
 * @class Limiter
 * @extends Decorator
**/
type Limiter struct {
	core.Decorator
	maxLoop int
}

// Initialize 初始化
// Settings 属性:
//
//	      key: maxLoop
//		  value: 循环次数
func (it *Limiter) Initialize() {
	it.Decorator.Initialize()
	// TODO: 设置循环限制次数
	//it.maxLoop = setting.GetPropertyAsInt("maxLoop")
	//if it.maxLoop < 1 {
	//	panic("maxLoop parameter in MaxTime decorator is an obligatory parameter")
	//}
}

func (it *Limiter) OnTick(tick *core.Tick) behavior.Status {
	if it.GetChild() == nil {
		return behavior.ERROR
	}
	var i = tick.Blackboard.GetInt("i", tick.GetTree().GetID(), it.GetID())
	if i < it.maxLoop {
		var status = it.GetChild().Execute(tick)
		if status == behavior.SUCCESS || status == behavior.FAILURE {
			tick.Blackboard.Set("i", i+1, tick.GetTree().GetID(), it.GetID())
		}
		return status
	}

	return behavior.FAILURE
}
