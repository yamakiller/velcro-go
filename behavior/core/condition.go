package core

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/datas"
)

type ICondition interface {
	IBaseNode
}

// Condition 根据条件判断返回success或failed
type Condition struct {
	BaseNode
	BaseWorker
}

// Ctor ...
func (cdt *Condition) Ctor() {

	cdt.category = behavior.CONDITION
}

// Initialize 初始化
func (cdt *Condition) Initialize(data *datas.Behavior3Node) {
	cdt.BaseNode.Initialize(data)
	//cdt.BaseNode.IBaseWorker = cdt
}
