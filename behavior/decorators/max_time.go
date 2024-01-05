package decorators

import (
	"time"

	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// 可以执行的最长时间.
// 请注意,它不会中断执行本身(即子进程必须是非抢占式的),它仅在"RUNNING"状态后中断节点.
type MaxTime struct {
	core.Decorator
	maxTime int64
}

func (mt *MaxTime) Initialize() {
	mt.Decorator.Initialize()
	//this.maxTime = setting.GetPropertyAsInt64("maxTime")
	//if this.maxTime < 1 {
	//	panic("maxTime parameter in Limiter decorator is an obligatory parameter")
	//}
}

func (mt *MaxTime) OnOpen(tick *core.Tick) {
	var startTime int64 = time.Now().UnixNano() / 1000000
	tick.Blackboard.Set("startTime", startTime, tick.GetTree().GetID(), mt.GetID())
}

func (mt *MaxTime) OnTick(tick *core.Tick) behavior.Status {
	if mt.GetChild() == nil {
		return behavior.ERROR
	}
	var currTime int64 = time.Now().UnixNano() / 1000000
	var startTime int64 = tick.Blackboard.GetInt64("startTime", tick.GetTree().GetID(), mt.GetID())
	var status = mt.GetChild().Execute(tick)
	if currTime-startTime > mt.maxTime {
		return behavior.FAILURE
	}

	return status
}
