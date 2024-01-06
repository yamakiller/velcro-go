package decorators

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// Succeeder 不论子节点返回结果是什么,都返回success.
type Succeeder struct {
	core.Decorator
}

func (s *Succeeder) OnTick(tick *core.Tick) behavior.Status {
	if s.GetChild() != nil {
		s.GetChild().Execute(tick)
	}
	return behavior.SUCCESS
}
