package decorators

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// AlwaysFailed 除了running,不论子节点返回结果是什么,都返回failed.
type AlwaysFailed struct {
	core.Decorator
}

func (alf *AlwaysFailed) OnTick(tick *core.Tick) behavior.Status {
	child := alf.GetChild()
	if child == nil {
		return behavior.FAILURE
	}
	status := child.Execute(tick)
	if status == behavior.RUNNING {
		return behavior.RUNNING
	}
	return behavior.FAILURE
}
