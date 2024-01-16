package actions

import (
	"time"

	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/datas"
)

// Wait 等待一段时间，然后返回Success
type Wait struct {
	core.Action
	duration int64
}

// 初始化方法
//
// Settings 属性:
//
//	      key: milliseconds
//		  value: 毫秒时间
func (wt *Wait) Initialize(data *datas.Behavior3Node) {
	wt.Action.Initialize(data)
	wt.duration = data.GetInt64("duration_millis")
}

// OnOpen 打开节点方法.
func (wt *Wait) OnOpen(tick *core.Tick) {
	var startTime int64 = time.Now().UnixNano() / 1000000
	tick.Blackboard.Set("startTime", startTime, tick.GetTree().GetID(), wt.GetID())
}

// Tick 方法.
func (wt *Wait) OnTick(tick *core.Tick) behavior.Status {
	var currTime int64 = time.Now().UnixNano() / 1000000
	var startTime = tick.Blackboard.GetInt64("startTime", tick.GetTree().GetID(), wt.GetID())

	if currTime-startTime > wt.duration {
		return behavior.SUCCESS
	}

	return behavior.RUNNING
}
