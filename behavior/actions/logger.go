package actions

import (
	"fmt"

	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
)

type Log struct {
	core.Action
	info string
}

func (lg *Log) Initialize() {
	lg.Action.Initialize()
	//this.info = setting.GetPropertyAsString("info")
}

func (lg *Log) OnTick(tick *core.Tick) behavior.Status {
	fmt.Println("log:", lg.info)
	return behavior.SUCCESS
}
