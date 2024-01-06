package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// ParallelSelector 每次都执行所有节点, 只要有一个子节点返回success, 那么
// 当前节点也返回success, 当所有节点都failure,当前节点也failure.
type ParallelSelector struct {
	core.Composite
}

func (pstor *ParallelSelector) Tick(tick *core.Tick) behavior.Status {
	resultStatus := behavior.FAILURE
	for i := 0; i < pstor.GetChildCount(); i++ {
		var status = pstor.GetChild(i).Execute(tick)
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
