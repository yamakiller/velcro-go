package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/datas"
	"github.com/yamakiller/velcro-go/utils/lang/fastrand"
)

// RandomSequence 随机顺序执行,随机将children排序,然后
// 执行,其它与Sequence相同.
type RandomSequence struct {
	core.Composite
	rndIndex []int
}

func (rseq *RandomSequence) Initialize(data *datas.Behavior3Node) {
	rseq.Composite.Initialize(data)
	rseq.rndIndex = make([]int, rseq.GetChildCount())
}

func (rseq *RandomSequence) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("runningChild", 0, tick.GetTree().GetID(), rseq.GetID())
}

func (rseq *RandomSequence) OnEnter(tick *core.Tick) {
	// 重新随机执行顺序
	for i := len(rseq.rndIndex) - 1; i > 0; i-- {
		j := fastrand.Intn(i + 1)
		rseq.rndIndex[i], rseq.rndIndex[j] = rseq.rndIndex[j], rseq.rndIndex[i]
	}
}

func (rseq *RandomSequence) OnTick(tick *core.Tick) behavior.Status {
	var child = tick.Blackboard.GetInt("runningChild", tick.GetTree().GetID(), rseq.GetID())
	for i := child; i < rseq.GetChildCount(); i++ {
		var status = rseq.GetChild(rseq.rndIndex[i]).Execute(tick)
		if status == behavior.RUNNING {
			// save ruuning state, restart from i next time
			tick.Blackboard.Set("runningChild", i, tick.GetTree().GetID(), rseq.GetID())

			return behavior.RUNNING
		} else if status == behavior.FAILURE {
			return behavior.FAILURE
		}
	}

	return behavior.SUCCESS
}
