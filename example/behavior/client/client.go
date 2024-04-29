package client

import (
	"time"

	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/loader"
	"github.com/yamakiller/velcro-go/behavior/registers"
	"github.com/yamakiller/velcro-go/example/behavior/kafka"
	"github.com/yamakiller/velcro-go/example/behavior/nodes"
	"github.com/yamakiller/velcro-go/vlog"
)

func BInt(path string) {
	registers.Register("WaitLog", &nodes.WaitLog{})
	registers.Register("Log", &nodes.Log{})

	tree, err := loader.LoadBehaviorTreeFromFile(path)
	if err != nil {
		vlog.Error("[BT] Load Behavior Tree fail %v", err)
		return
	}
	core.SetSubTreeLoadFunc(tree.SelectBehaviorTree)
	selectedtree := tree.SelectedBehaviorDefaultTree()
	b := core.NewBlackboard()
	b.Set("CurrStreeNode", "", selectedtree.GetID(), "")
	t1 := time.NewTicker(1 * time.Second).C
	for {
		select {
		case <-t1:
			selectedtree.Tick(0, b)
			kafka.WriteCurrTree(selectedtree.GetID(), b.Get("CurrStreeNode", selectedtree.GetID(), "").(string))
		}
	}
	for i := 0; i >= 0; i++ {
		selectedtree.Tick(i, b)

	}
}
