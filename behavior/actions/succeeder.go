package actions

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

type Succeeder struct {
	core.Action
}

func (s *Succeeder) OnTick(tick *core.Tick) behavior.Status {
	return behavior.SUCCESS
}
