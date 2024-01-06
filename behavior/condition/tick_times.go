package condition

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/datas"
)

type TickTimes struct {
	core.Condition
	times int
}

func (tt *TickTimes) Initialize(data *datas.Behavior3Node) {
	tt.Condition.Initialize(data)
	tt.times = data.GetInt("times")
}

func (tt *TickTimes) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("runningTime", 0, tick.GetTree().GetID(), tt.GetID())
}

func (tt *TickTimes) OnTick(tick *core.Tick) behavior.Status {

	var runningTime = tick.Blackboard.GetInt("runningTime", tick.GetTree().GetID(), tt.GetID())
	runningTime++
	tick.Blackboard.Set("runningTime", runningTime, tick.GetTree().GetID(), tt.GetID())
	if runningTime >= tt.times {
		tick.Blackboard.Set("runningTime", 0, tick.GetTree().GetID(), tt.GetID())
		return behavior.SUCCESS
	}

	return behavior.FAILURE
}
