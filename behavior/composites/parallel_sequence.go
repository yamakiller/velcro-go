package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// ParallelSequence 每次都会执行所有节点, 只要有一个子节点返回failure, 那么
// 当前节点也返回failure,当所有节点都成功, 当前节点也返回成功.
type ParallelSequence struct {
	core.Composite
}

func (pseq *ParallelSequence) Tick(tick *core.Tick) behavior.Status {
	resultStatus := behavior.SUCCESS
	for i := 0; i < pseq.GetChildCount(); i++ {
		var status = pseq.GetChild(i).Execute(tick)
		switch status {
		case behavior.RUNNING:
			if resultStatus == behavior.SUCCESS {
				resultStatus = behavior.RUNNING
			}
		case behavior.FAILURE:
			resultStatus = behavior.FAILURE
		}
	}
	return resultStatus
}
