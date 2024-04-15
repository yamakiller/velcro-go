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

func (this * WaitLog) Initialize(data *datas.Behavior3Node) {
	this.Decorator.Initialize(data)
	this.duration = data.GetInt64("duration_millis")
}

func (this * WaitLog) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("WaitLog",int64(0), tick.GetTree().GetID(), this.GetID())
	var startTime int64 = time.Now().UnixNano() / 1000000
	tick.Blackboard.Set("startTime", startTime, tick.GetTree().GetID(), this.GetID())
}
func (this * WaitLog) OnTick(tick *core.Tick) behavior.Status{
	var state = tick.Blackboard.GetInt64("WaitLog", tick.GetTree().GetID(), this.GetID())
	var currTime int64 = time.Now().UnixNano() / 1000000
	var startTime = tick.Blackboard.GetInt64("startTime", tick.GetTree().GetID(), this.GetID())
	if currTime-startTime > this.duration && state == 0{
		tick.Blackboard.Set("WaitLog", int64(1), tick.GetTree().GetID(), this.GetID())
		tick.Blackboard.Set("Log",fmt.Sprintf("stree %s | node %s ",tick.GetTree().GetID(), this.GetID()),tick.GetTree().GetID(),"")
		return this.GetChild().Execute(tick)
	}
	return behavior.RUNNING
}