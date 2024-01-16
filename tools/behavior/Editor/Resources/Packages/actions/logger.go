package actions

import (
	"fmt"

	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/datas"
)

type Log struct {
	core.Action
	info string
}

func (lg *Log) Initialize(data *datas.Behavior3Node) {
	lg.Action.Initialize(data)
	lg.info = data.GetString("info")
}

func (lg *Log) OnTick(tick *core.Tick) behavior.Status {
	fmt.Println("log:", lg.info)
	return behavior.SUCCESS
}
