package decorators

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/datas"
)

// UntilFailure 重复执行子节点，直到子节点返回failed,此节点会返回success.或者达到maxLoop次后,此节点返回failed.
type UntilFailure struct {
	core.Decorator
	maxLoop int
}

func (uf *UntilFailure) Initialize(data *datas.Behavior3Node) {
	uf.Decorator.Initialize(data)
	uf.maxLoop = data.GetInt("maxLoop")
	if uf.maxLoop < 1 {
		panic("maxLoop parameter in MaxTime decorator is an obligatory parameter")
	}
}

func (uf *UntilFailure) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("times", 0, tick.GetTree().GetID(), uf.GetID())
}

func (uf *UntilFailure) OnTick(tick *core.Tick) behavior.Status {
	if uf.GetChild() == nil {
		return behavior.ERROR
	}

	var times = tick.Blackboard.GetInt("times", tick.GetTree().GetID(), uf.GetID())
	if uf.maxLoop <= 0 || times < uf.maxLoop {
		var status = uf.GetChild().Execute(tick)
		times++
		tick.Blackboard.Set("times", times, tick.GetTree().GetID(), uf.GetID())
		if status == behavior.FAILURE {
			return behavior.SUCCESS
		}
		return behavior.RUNNING
	}
	return behavior.FAILURE
}
