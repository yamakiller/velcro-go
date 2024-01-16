package actions

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

type Runner struct {
	core.Action
}

func (r *Runner) OnTick(tick *core.Tick) behavior.Status {
	return behavior.RUNNING
}
