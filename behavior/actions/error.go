package actions

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

type Error struct {
	core.Action
}

func (e *Error) OnTick(tick *core.Tick) behavior.Status {
	return behavior.ERROR
}
