package core

import (
	"github.com/yamakiller/velcro-go/behavior"
)

type IAction interface {
	IBaseNode
}

// Action 是所有操作节点的基类.因此,如果要创建新的自定义操作节点,则需要继承此类.例如,看一下 Runner 动作:
//
//	var Runner = behavior.Class(b3.Action, {
//	    name: 'Runner',
//
//	    tick: function(tick) {
//	      return b3.RUNNING;
//	    }
//	  });
type Action struct {
	BaseNode
	BaseWorker
}

func (at *Action) Ctor() {
	at.category = behavior.ACTION
}
func (at *Action) Initialize() {

	//at.id = shortuuid.New()
	at.BaseNode.Initialize()
	//at.BaseNode.IBaseWorker = this
	at.properties = make(map[string]interface{})
}
