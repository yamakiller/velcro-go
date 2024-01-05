package core

import "github.com/yamakiller/velcro-go/behavior"

type ICondition interface {
	IBaseNode
}

type Condition struct {
	BaseNode
	BaseWorker
}

// Ctor ...
func (cdt *Condition) Ctor() {

	cdt.category = behavior.CONDITION
}

// Initialize 初始化
func (cdt *Condition) Initialize() {
	cdt.BaseNode.Initialize()
	//cdt.BaseNode.IBaseWorker = cdt
}
