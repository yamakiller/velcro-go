package core

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/datas"
)

type SubTree struct {
	Action
}

// Initialize 初始化子树木
func (sbt *SubTree) Initialize(data *datas.Behavior3Node) {
	sbt.Action.Initialize(data)
}

// OnTick 执行子树
// 使用sTree.Tick(tar, tick.Blackboard)的方法会导致每个树有自己的tick.
// 如果子树包含running状态，同时复用了子树会导致歧义.
// 改为只使用一个树，一个tick上下文.
func (sbt *SubTree) OnTick(tick *Tick) behavior.Status {

	//使用子树，必须先SetSubTreeLoadFunc
	//子树可能没有加载上来，所以要延迟加载执行
	sTree := subTreeLoadFunc(sbt.GetName())
	if nil == sTree {
		return behavior.ERROR
	}

	if tick.GetTarget() == nil {
		panic("SubTree tick.GetTarget() nil !")
	}

	//tar := tick.GetTarget()
	//return sTree.Tick(tar, tick.Blackboard)

	tick.pushSubtreeNode(sbt)
	ret := sTree.GetRoot().Execute(tick)
	tick.popSubtreeNode()
	return ret
}

func (sbt *SubTree) String() string {
	return "SBT_" + sbt.GetTitle()
}

var subTreeLoadFunc func(string) *BehaviorTree

// 获取子树的方法
func SetSubTreeLoadFunc(f func(string) *BehaviorTree) {
	subTreeLoadFunc = f
}
