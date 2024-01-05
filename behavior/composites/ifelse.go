package composites

import "github.com/yamakiller/velcro-go/behavior/core"

type Ifels struct {
	core.Composite
}

func (els *Ifels) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("runningChild", 0, tick.GetTree().GetID(), els.GetID())
	if els.GetChildCount() != 3 {
		panic("IfElseTask has to have three children: condition, if, else")
	}
}
