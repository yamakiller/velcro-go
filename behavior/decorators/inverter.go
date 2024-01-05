package decorators

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// Inverter 装饰器节点
//   - 反转子级的结果.
//     为"FAILURE"返回"SUCCESS", 为"SUCCESS"返回"FAILURE".
type Inverter struct {
	core.Decorator
}

func (inv *Inverter) OnTick(tick *core.Tick) behavior.Status {
	if inv.GetChild() == nil {
		return behavior.ERROR
	}

	var status = inv.GetChild().Execute(tick)
	if status == behavior.SUCCESS {
		status = behavior.FAILURE
	} else if status == behavior.FAILURE {
		status = behavior.SUCCESS
	}

	return status
}
