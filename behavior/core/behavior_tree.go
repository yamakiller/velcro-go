package core

import (
	"fmt"

	"github.com/lithammer/shortuuid/v4"
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/datas"
	"github.com/yamakiller/velcro-go/behavior/registers"
)

type BehaviorTree struct {
	// 树ID(只读)
	id string
	// 树标题(只读)
	title string
	// 树描述(只读)
	description string
	//(只读)
	properties map[string]interface{}
	// 跟节点
	root IBaseNode
	// 调试接口
	debug interface{}
}

// Initialize 初始化
func (bb *BehaviorTree) Initialize() {
	bb.id = shortuuid.New()
	bb.title = "The behavior tree"
	bb.description = "Default description"
	bb.properties = make(map[string]interface{})
	bb.root = nil
	bb.debug = nil
}

func (bb *BehaviorTree) GetID() string {
	return bb.id
}

func (bb *BehaviorTree) GetTitile() string {
	return bb.title
}

func (bb *BehaviorTree) SetDebug(debug interface{}) {
	bb.debug = debug
}

func (bb *BehaviorTree) GetRoot() IBaseNode {
	return bb.root
}

func (bb *BehaviorTree) Load(data *datas.Behavior3Tree) {
	bb.title = data.Title
	bb.description = data.Description
	bb.properties = data.Properties
	nodes := make(map[string]IBaseNode)

	for id, s := range data.Nodes {
		spec := s
		var node IBaseNode
		if spec.Category == "tree" {
			node = new(SubTree)
		} else {
			if tnode, err := registers.New(spec.Name); err != nil {
				node = tnode.(IBaseNode)
			}
		}

		if node == nil {
			panic("BehaviorTree.load: Invalid node name:" + spec.Name + ",title:" + spec.Title)
		}

		node.Ctor()
		node.Initialize(spec)
		node.SetBaseNodeWorker(node.(IBaseWorker))
		nodes[id] = node
	}

	// Connect the nodes
	for id, spec := range data.Nodes {
		node := nodes[id]

		if node.GetCategory() == behavior.COMPOSITE && spec.Children != nil {
			for i := 0; i < len(spec.Children); i++ {
				var cid = spec.Children[i]
				comp := node.(IComposite)
				if nodes[cid] == nil {
					panic("BehaviorTree.load: Invalid node id: " + cid)
				}
				comp.AddChild(nodes[cid])
			}
		} else if node.GetCategory() == behavior.DECORATOR && len(spec.Child) > 0 {
			dec := node.(IDecorator)
			dec.SetChild(nodes[spec.Child])
		}
	}
	bb.root = nodes[data.Root]
}

// 从根部开始在树中传播刻度信号
//
// 此方法接收任何类型的目标对象(Object、Array、DOMElement 等)和"Blackboard"实例.
// 目标对象对于所有 Behavior3JS 组件来说根本没有用处，但对于自定义节点来说肯定很重要.
// Blackboard 实例被树和节点用来存储执行变量(例如,最后运行的节点),并且必须是"Blackboard"
// 实例(或具有相同接口的对象).
//
// 在内部,此方法创建一个 Tick 对象,该对象将存储目标和黑板对象.
//
// Note: BehaviourTree 存储上一个tick 以来打开的节点列表, 如果这些节点在当前tick 之后没有被调用, 此方
// 法将自动关闭它们.
//
// blackboard 黑板对象的实例.
// 返回信号状态
func (bt *BehaviorTree) Tick(target interface{}, blackboard *Blackboard) behavior.Status {
	if blackboard == nil {
		panic("The blackboard parameter is obligatory and must be an instance of b3.Blackboard")
	}

	// CREATE A TICK OBJECT
	var tick = NewTick()
	tick.debug = bt.debug
	tick.target = target
	tick.Blackboard = blackboard
	tick.tree = bt

	// TICK NODE
	var state = bt.root.execute(tick)

	// CLOSE NODES FROM LAST TICK, IF NEEDED
	var lastOpenNodes = blackboard.getTreeData(bt.id).OpenNodes
	var currOpenNodes []IBaseNode
	currOpenNodes = append(currOpenNodes, tick.openNodes...)

	// does not close if it is still open in this tick
	var start = 0
	for i := 0; i < behavior.MinInt(len(lastOpenNodes), len(currOpenNodes)); i++ {
		start = i + 1
		if lastOpenNodes[i] != currOpenNodes[i] {
			break
		}
	}

	// close the nodes
	for i := len(lastOpenNodes) - 1; i >= start; i-- {
		lastOpenNodes[i].close(tick)
	}

	// POPULATE BLACKBOARD
	blackboard.getTreeData(bt.id).OpenNodes = currOpenNodes
	blackboard.SetTree("nodeCount", tick.nodeCount, bt.id)

	return state
}

func (bt *BehaviorTree) Print() {
	printNode(bt.root, 0)
}

func printNode(root IBaseNode, blk int) {

	//fmt.Println("new node:", root.Name, " children:", len(root.Children), " child:", root.Child)
	for i := 0; i < blk; i++ {
		fmt.Print(" ") //缩进
	}

	//fmt.Println("|—<", root.Name, ">") //打印"|—<id>"形式
	fmt.Print("|—", root.GetTitle())

	if root.GetCategory() == behavior.DECORATOR {
		dec := root.(IDecorator)
		if dec.GetChild() != nil {
			//fmt.Print("=>")
			printNode(dec.GetChild(), blk+3)
		}
	}

	fmt.Println("")
	if root.GetCategory() == behavior.COMPOSITE {
		comp := root.(IComposite)
		if comp.GetChildCount() > 0 {
			for i := 0; i < comp.GetChildCount(); i++ {
				printNode(comp.GetChild(i), blk+3)
			}
		}
	}

}
