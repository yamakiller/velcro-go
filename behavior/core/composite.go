package core

import (
	"fmt"

	"github.com/yamakiller/velcro-go/behavior"
)

type IComposite interface {
	IBaseNode
	GetChildCount() int
	GetChild(index int) IBaseNode
	AddChild(child IBaseNode)
}

type Composite struct {
	BaseNode
	BaseWorker

	children []IBaseNode
}

func (cps *Composite) Ctor() {

	cps.category = behavior.COMPOSITE
}

// Initialize 初始化
func (cps *Composite) Initialize() {
	cps.BaseNode.Initialize()
	//cps.BaseNode.IBaseWorker = cps
	cps.children = make([]IBaseNode, 0)
	//fmt.Println("Composite Initialize")
}

// GetChildCount 返回子节点数量
func (cps *Composite) GetChildCount() int {
	return len(cps.children)
}

// GetChild 获取子节点
func (cps *Composite) GetChild(index int) IBaseNode {
	return cps.children[index]
}

// AddChild 添加子节点
func (cps *Composite) AddChild(child IBaseNode) {
	cps.children = append(cps.children, child)
}

func (cps *Composite) tick(tick *Tick) behavior.Status {
	fmt.Println("tick Composite1")
	return behavior.ERROR
}
