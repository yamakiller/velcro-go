package composites

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// Sequence 顺序执行，只要一个不成功就返回
type Sequence struct {
	core.Composite
}

// OnTick
func (seq *Sequence) OnTick(tick *core.Tick) behavior.Status {
	//fmt.Println("tick Sequence :", this.GetTitle())
	for i := 0; i < seq.GetChildCount(); i++ {
		var status = seq.GetChild(i).Execute(tick)
		if status != behavior.SUCCESS {
			return status
		}
	}
	return behavior.SUCCESS
}
