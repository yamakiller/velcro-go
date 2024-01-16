package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// Parallel 每次都执行所有节点,永远返回 running
type Parallel struct {
	core.Composite
}

func (pal *Parallel) OnTick(tick *core.Tick) behavior.Status {
	resultStatus := behavior.FAILURE
	for i := 0; i < pal.GetChildCount(); i++ {
		var status = pal.GetChild(i).Execute(tick)
		switch status {
		case behavior.RUNNING:
			if resultStatus == behavior.FAILURE {
				resultStatus = behavior.RUNNING
			}
		case behavior.SUCCESS:
			resultStatus = behavior.SUCCESS
		}
	}

	return resultStatus
}
