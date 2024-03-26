package client

import (
	"fmt"
	"os"

	"github.com/yamakiller/velcro-go/behavior/core"
	"github.com/yamakiller/velcro-go/behavior/loader"
	"github.com/yamakiller/velcro-go/behavior/registers"
	"github.com/yamakiller/velcro-go/example/behavior/nodes"
	"github.com/yamakiller/velcro-go/vlog"
)

func BInt(path string) {
	registers.Register("WaitLog",&nodes.WaitLog{})
	registers.Register("Log",&nodes.Log{})
	tree,err := loader.LoadBehaviorTreeFromFile(path)
	if err != nil{
		vlog.Error("[BT] Load Behavior Tree fail %v",err)
		return
	}
	core.SetSubTreeLoadFunc(tree.SelectBehaviorTree)
	selectedtree := tree.SelectedBehaviorDefaultTree()
	b := core.NewBlackboard()
	b.Set("CurrStreeNode","",selectedtree.GetID(),"")
	sss := ""
	for i:= 0; i < 100000000; i ++{
		selectedtree.Tick(i,b)
		b:= b.Get("CurrStreeNode",selectedtree.GetID(),"").(string)
		if b != sss{
			sss = b
			fmt.Fprintf(os.Stderr, "tree %s, node %s \n",selectedtree.GetID(), sss)	
		}
	}
}