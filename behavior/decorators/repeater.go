package decorators

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

// Repeater是一个装饰器, 它会重复tick信号, 直到子节点返回
// 'RUNNING'或'ERROR'.或者,可以定义最大重复次数.
type Repeater struct {
	core.Decorator
	maxLoop int
}

// Initialize 初始化
// Settings 属性:
//
//	      key: maxLoop
//		  value: 循环次数
func (rp *Repeater) Initialize() {
	rp.Decorator.Initialize()
	//this.maxLoop = setting.GetPropertyAsInt("maxLoop")
	//if this.maxLoop < 1 {
	//	panic("maxLoop parameter in MaxTime decorator is an obligatory parameter")
	//}
}

func (rp *Repeater) OnOpen(tick *core.Tick) {
	tick.Blackboard.Set("i", 0, tick.GetTree().GetID(), rp.GetID())
}

func (rp *Repeater) OnTick(tick *core.Tick) behavior.Status {
	//fmt.Println("tick ", this.GetTitle())
	if rp.GetChild() == nil {
		return behavior.ERROR
	}
	var i = tick.Blackboard.GetInt("i", tick.GetTree().GetID(), rp.GetID())
	var status = behavior.SUCCESS
	for rp.maxLoop < 0 || i < rp.maxLoop {
		status = rp.GetChild().Execute(tick)
		if status == behavior.SUCCESS || status == behavior.FAILURE {
			i++
		} else {
			break
		}
	}
	tick.Blackboard.Set("i", i, tick.GetTree().GetID(), rp.GetID())
	return status
}
