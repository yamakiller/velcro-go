package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

type Ifels struct {
	core.Composite
}

func (els *Ifels) OnOpen(tick *core.Tick) {
	if els.GetChildCount() != 3 {
		panic("IfElseTask has to have three children: condition, if, else")
	}
}

func (els *Ifels) OnTick(tick *core.Tick) behavior.Status {
	var status = els.GetChild(0).Execute(tick)
	if status == behavior.SUCCESS {
		return els.GetChild(1).Execute(tick)
	}

	return els.GetChild(2).Execute(tick)
}
