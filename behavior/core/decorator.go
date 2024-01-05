package core

import "github.com/yamakiller/velcro-go/behavior"

type IDecorator interface {
	IBaseNode
	SetChild(child IBaseNode)
	GetChild() IBaseNode
}

type Decorator struct {
	BaseNode
	BaseWorker
	child IBaseNode
}

// Ctor ...
func (d *Decorator) Ctor() {
	d.category = behavior.DECORATOR
}

// Initialize 初始化方法
func (d *Decorator) Initialize() {
	d.BaseNode.Initialize()
	//this.BaseNode.IBaseWorker = this
}

// GetChild 返回字节点
func (d *Decorator) GetChild() IBaseNode {
	return d.child
}

// SetChild 设置子节点
func (d *Decorator) SetChild(child IBaseNode) {
	d.child = child
}
