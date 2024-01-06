package decorators

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// AlwaysSuccess 除了running, 不论子节点返回结果是什么,都返回success.
type AlwaysSuccess struct {
	core.Decorator
}

func (als *AlwaysSuccess) OnTick(tick *core.Tick) behavior.Status {
	child := als.GetChild()
	if child == nil {
		return behavior.FAILURE
	}
	status := child.Execute(tick)
	if status == behavior.RUNNING {
		return behavior.RUNNING
	}
	return behavior.SUCCESS
}
