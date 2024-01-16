package actions

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

type Failer struct {
	core.Action
}

func (f *Failer) OnTick(tick *core.Tick) behavior.Status {
	return behavior.FAILURE
}
