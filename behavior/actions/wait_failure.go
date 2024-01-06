package actions

import (
	"time"

	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/datas"
)

// WaitFailure 等待一段时间，然后返回Failure.
type WaitFailure struct {
	core.Action
	duration int64
}

func (wt *WaitFailure) Initialize(data *datas.Behavior3Node) {
	wt.Action.Initialize(data)
	wt.duration = data.GetInt64("duration")
}

// OnOpen 打开节点方法.
func (wt *WaitFailure) OnOpen(tick *core.Tick) {
	var startTime int64 = time.Now().UnixNano() / 1000000
	tick.Blackboard.Set("startTime", startTime, tick.GetTree().GetID(), wt.GetID())
}

func (wt *WaitFailure) OnTick(tick *core.Tick) behavior.Status {
	var currTime int64 = time.Now().UnixNano() / 1000000
	var startTime = tick.Blackboard.GetInt64("startTime", tick.GetTree().GetID(), wt.GetID())

	if currTime-startTime > wt.duration {
		return behavior.FAILURE
	}

	return behavior.RUNNING
}
