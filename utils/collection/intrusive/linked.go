package intrusive

import "sync"

func NewLinked(mutex sync.Locker) *Linked {
	return &Linked{mutex: mutex, head: nil, tail: nil}
}

type INode interface {
	Prev() INode
	Next() INode
	WithPrev(INode)
	WithNext(INode)
}

type LinkedNode struct {
	prev INode
	next INode
}

func (n *LinkedNode) Prev() INode {
	return n.prev
}

func (n *LinkedNode) Next() INode {
	return n.next
}

func (n *LinkedNode) WithPrev(node INode) {
	n.prev = node
}

func (n *LinkedNode) WithNext(node INode) {
	n.next = node
}

type Linked struct {
	head  INode
	tail  INode
	mutex sync.Locker
}

// Push 尾部加入一条数据并返回这个节点
func (linked *Linked) Push(node INode) {
	linked.mutex.Lock()
	defer linked.mutex.Unlock()

	node.WithNext(nil)
	node.WithPrev(nil)
	if linked.head == nil {
		linked.head = node
		linked.tail = node
	} else {
		node.WithPrev(linked.tail)
		linked.tail.WithNext(node)
		linked.tail = node
	}
}

// Pop 弹出尾部的节点
func (linked *Linked) Pop() INode {
	linked.mutex.Lock()
	defer linked.mutex.Unlock()

	if linked.tail == nil {
		return nil
	}
	popNode := linked.tail

	if popNode.Prev() != nil {
		popNode.Prev().WithNext(nil)
		linked.tail = popNode.Prev()
	}else{
		linked.head = nil
		linked.tail = nil
	}
	
	popNode.WithPrev(nil)
	popNode.WithNext(nil)
	return popNode
}

// Remove 删除这个节点
func (linked *Linked) Remove(node INode) {
	linked.mutex.Lock()
	defer linked.mutex.Unlock()
	if node == nil{
		return
	}
	if node.Prev() != nil {
		node.Prev().WithNext(node.Next())
	}

	if node.Next() != nil {
		node.Next().WithPrev(node.Prev())
	}

	if node == linked.head {
		linked.head = nil
		linked.tail = nil
	}

	if node == linked.tail && node.Prev() != nil {
		linked.tail = node.Prev()
	}

	node.WithPrev(nil)
	node.WithNext(nil)
}

func (linked *Linked) Foreach(f func(INode)bool){
	if f == nil{
		return
	}
	linked.mutex.Lock()
	defer linked.mutex.Unlock()
	if linked.head == nil {
		return
	}

	tmp := linked.head
	for tmp != nil{
		node := tmp.Next()
		f(tmp)
		tmp = node
	}
}

func (linked *Linked) Len() int {
	linked.mutex.Lock()
	defer linked.mutex.Unlock()
	p := linked.head
	len := 0
	for p != nil{
		len++
		if p == linked.tail{
			return len
		}
		p = p.Next()
	}

	return len
}