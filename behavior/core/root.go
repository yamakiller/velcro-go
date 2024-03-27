package core

import (
	"fmt"

	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/datas"
)

type IRoot interface {
	IBaseNode
	SetChild(child IBaseNode)
	GetChild() IBaseNode
}

type Root struct {
	BaseNode
	BaseWorker

	child IBaseNode
}

func (rt *Root) Ctor() {
	rt.category = behavior.Root
}

// Initialize 初始化
func (rt *Root) Initialize(data *datas.Behavior3Node) {
	rt.BaseNode.Initialize(data)
	//rt.BaseNode.IBaseWorker = cps
	// rt.children = make([]IBaseNode, 0)
	//fmt.Println("Composite Initialize")
}

// GetChild 返回字节点
func (rt *Root) GetChild() IBaseNode {
	return rt.child
}

// SetChild 设置子节点
func (rt *Root) SetChild(child IBaseNode) {
	rt.child = child
}

func (rt *Root) OnTick(tick *Tick) behavior.Status{
	if rt.child != nil{
		return  rt.child.Execute(tick)
	}
	return behavior.SUCCESS
}

func (rt *Root) tick(tick *Tick) behavior.Status {
	fmt.Println("tick Composite1")
	return behavior.ERROR
}
