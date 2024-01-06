package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/datas"
	"github.com/yamakiller/velcro-go/utils/lang/fastrand"
)

// RandomSelector 随机顺序执行,随机将children排序,然后执行,其它
// 与RandomSelector相同
type RandomSelector struct {
	core.Composite
	rndIndex []int
}

func (rstor *RandomSelector) Initialize(data *datas.Behavior3Node) {
	rstor.Composite.Initialize(data)
	rstor.rndIndex = make([]int, rstor.GetChildCount())
}

func (rstor *RandomSelector) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("runningChild", 0, tick.GetTree().GetID(), rstor.GetID())
}

func (rstor *RandomSelector) OnEnter(tick *core.Tick) {
	// 重新随机执行顺序
	for i := len(rstor.rndIndex) - 1; i > 0; i-- {
		j := fastrand.Intn(i + 1)
		rstor.rndIndex[i], rstor.rndIndex[j] = rstor.rndIndex[j], rstor.rndIndex[i]
	}
}

func (rstor *RandomSelector) OnTick(tick *core.Tick) behavior.Status {
	var child = tick.Blackboard.GetInt("runningChild", tick.GetTree().GetID(), rstor.GetID())
	for i := child; i < rstor.GetChildCount(); i++ {
		var status = rstor.GetChild(rstor.rndIndex[i]).Execute(tick)
		if status == behavior.RUNNING {
			tick.Blackboard.Set("runningChild", i, tick.GetTree().GetID(), rstor.GetID())
			return behavior.RUNNING
		} else if status == behavior.SUCCESS {
			return behavior.SUCCESS
		}
	}
	return behavior.FAILURE
}
