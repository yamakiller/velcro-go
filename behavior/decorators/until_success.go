package decorators

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/datas"
)

// UntilSuccess 重复执行子节点,直到子节点返回success,此节点会返回success.
// 或者达到maxLoop次后，此节点返回failed.
type UntilSuccess struct {
	core.Decorator
	maxLoop int
}

func (us *UntilSuccess) Initialize(data *datas.Behavior3Node) {
	us.Decorator.Initialize(data)
	us.maxLoop = data.GetInt("maxLoop")
	if us.maxLoop < 1 {
		panic("maxLoop parameter in maxLoop decorator is an obligatory parameter")
	}
}

func (us *UntilSuccess) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("loopTime", 0, tick.GetTree().GetID(), us.GetID())
}

func (us *UntilSuccess) OnTick(tick *core.Tick) behavior.Status {
	if us.GetChild() == nil {
		return behavior.ERROR
	}

	var loopTime = tick.Blackboard.GetInt("loopTime", tick.GetTree().GetID(), us.GetID())
	var status = behavior.ERROR
	if us.maxLoop < 0 || loopTime < us.maxLoop {
		status = us.GetChild().Execute(tick)
		loopTime++
		tick.Blackboard.Set("loopTime", loopTime, tick.GetTree().GetID(), us.GetID())
		if status == behavior.FAILURE {
			return behavior.SUCCESS
		}
		return behavior.RUNNING
	}

	return behavior.FAILURE
}
