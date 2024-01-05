package core

func NewTick() *Tick {
	tick := &Tick{}
	tick.Initialize()
	return tick
}

type Tick struct {
	tree             *BehaviorTree
	debug            interface{}
	target           interface{}
	Blackboard       *Blackboard
	openNodes        []IBaseNode
	openSubtreeNodes []*SubTree
	nodeCount        int
}

// Initialize 初始化Tick
func (t *Tick) Initialize() {
	// set by BehaviorTree
	t.tree = nil
	t.debug = nil
	t.target = nil
	t.Blackboard = nil

	// updated during the tick signal
	t.openNodes = nil
	t.openSubtreeNodes = nil
	t.nodeCount = 0
}

func (t *Tick) GetTree() *BehaviorTree {
	return t.tree
}

func (t *Tick) enterNode(node IBaseNode) {
	t.nodeCount++
	t.openNodes = append(t.openNodes, node)

	// TODO: call debug here
}

func (t *Tick) openNode(node *BaseNode) {
	// TODO: call debug here
}

// 勾选节点时的回调(由BaseNode调用)
func (t *Tick) tickNode(node *BaseNode) {
	// TODO: call debug here
	//fmt.Println("Tick _tickNode :", t.debug, " id:", node.GetID(), node.GetTitle())
}

func (t *Tick) closeNode(node *BaseNode) {
	// TODO: call debug here

	ulen := len(t.openNodes)
	if ulen > 0 {
		t.openNodes = t.openNodes[:ulen-1]
	}

}

func (t *Tick) pushSubtreeNode(node *SubTree) {
	t.openSubtreeNodes = append(t.openSubtreeNodes, node)
}

func (t *Tick) popSubtreeNode() {
	ulen := len(t.openSubtreeNodes)
	if ulen > 0 {
		t.openSubtreeNodes = t.openSubtreeNodes[:ulen-1]
	}
}

// GetLastSubTree 返回Top子树节点.
// nil when it is runing at major tree
func (t *Tick) GetLastSubTree() *SubTree {
	ulen := len(t.openSubtreeNodes)
	if ulen > 0 {
		return t.openSubtreeNodes[ulen-1]
	}
	return nil
}

// 退出节点时的回调(由BaseNode调用).
// node 调用该方法的节点
func (t *Tick) exitNode(node *BaseNode) {
	// TODO: call debug here
}

func (t *Tick) GetTarget() interface{} {
	return t.target
}
