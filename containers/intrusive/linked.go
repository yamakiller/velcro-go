package intrusive

import "sync"

func NewLinked(mutex sync.Locker) *Linked {
	return &Linked{mutex: mutex, head: nil, tail: nil}
}

type LinkedNode struct {
	prev  *LinkedNode
	next  *LinkedNode
	Value interface{}
}

type Linked struct {
	head  *LinkedNode
	tail  *LinkedNode
	mutex sync.Locker
}

// Push 尾部加入一条数据并返回这个节点
func (linked *Linked) Push(value interface{}) *LinkedNode {
	linked.mutex.Lock()
	defer linked.mutex.Unlock()

	newNode := &LinkedNode{Value: value}
	if linked.head == nil {
		linked.head = newNode
		linked.tail = newNode
	} else {
		newNode.prev = linked.tail
		linked.tail.next = newNode
		linked.tail = newNode
	}

	return newNode
}

// Pop 弹出尾部的节点
func (linked *Linked) Pop() *LinkedNode {
	linked.mutex.Lock()
	defer linked.mutex.Unlock()

	if linked.tail == nil {
		return nil
	}

	popNode := linked.tail

	if popNode.prev != nil {
		popNode.prev.next = nil
	}

	linked.tail = popNode.prev
	popNode.prev = nil

	return popNode
}

// Remove 删除这个节点
func (linked *Linked) Remove(node *LinkedNode) {
	linked.mutex.Lock()
	defer linked.mutex.Unlock()

	if node.prev != nil {
		node.prev.next = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	}

	if node == linked.head {
		linked.head = node.next
	}

	if node == linked.tail {
		linked.tail = node.prev
	}

	node.prev = nil
	node.next = nil
}
