package actions

import (
	"time"

	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/datas"
	"github.com/yamakiller/velcro-go/utils/lang/fastrand"
)

// WaitRandom 随机等待一段时间，然后返回Success
type WaitRandom struct {
	core.Action
	rndDuration int64
	minDuration int64
}

func (wt *WaitRandom) Initialize(data *datas.Behavior3Node) {
	wt.Action.Initialize(data)
	max := data.GetInt64("max_duration_millis")
	wt.minDuration = data.GetInt64("min_duration_millis")
	wt.rndDuration = max - wt.minDuration
}

// OnOpen 打开节点方法.
func (wt *WaitRandom) OnOpen(tick *core.Tick) {
	var nowTime int64 = time.Now().UnixNano() / 1000000
	dur := fastrand.Int63n(wt.rndDuration) + wt.minDuration
	tick.Blackboard.Set("endTime", nowTime+dur, tick.GetTree().GetID(), wt.GetID())
}

func (wt *WaitRandom) OnTick(tick *core.Tick) behavior.Status {
	var currTime int64 = time.Now().UnixNano() / 1000000

	if currTime > tick.Blackboard.GetInt64("endTime", tick.GetTree().GetID(), wt.GetID()) {
		return behavior.SUCCESS
	}

	return behavior.RUNNING
}
