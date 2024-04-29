package nodes

import (
	"fmt"
	"time"

	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/datas"
)

type WaitLog struct {
	core.Decorator
	duration int64
}

func (w *WaitLog) Initialize(data *datas.Behavior3Node) {
	w.Decorator.Initialize(data)
	w.duration = data.GetInt64("duration_millis")
}

func (w *WaitLog) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("WaitLog", int64(0), tick.GetTree().GetID(), w.GetID())
	var startTime int64 = time.Now().UnixNano() / 1000000
	tick.Blackboard.Set("startTime", startTime, tick.GetTree().GetID(), w.GetID())
}
func (w *WaitLog) OnTick(tick *core.Tick) behavior.Status {
	var state = tick.Blackboard.GetInt64("WaitLog", tick.GetTree().GetID(), w.GetID())
	var currTime int64 = time.Now().UnixNano() / 1000000
	var startTime = tick.Blackboard.GetInt64("startTime", tick.GetTree().GetID(), w.GetID())
	if currTime-startTime > w.duration && state == 0 {
		tick.Blackboard.Set("WaitLog", int64(1), tick.GetTree().GetID(), w.GetID())
		tick.Blackboard.Set("Log", fmt.Sprintf("stree %s | node %s ", tick.GetTree().GetID(), w.GetID()), tick.GetTree().GetID(), "")
		return w.GetChild().Execute(tick)
	}
	return behavior.RUNNING
}
