// 黑板
package core

import (
	"fmt"
	"reflect"
)

// Blackboard 是"BehaviorTree"及其节点所需的内存结构.它只有 2 个公共方法:"set"和"get".
// 这些方法在 3 种不同的上下文中工作: 全局、每棵树和每棵树的每个节点.
//
// 假设您有两个不同的树用一个黑板控制单个对象, 那么:
//
//   - 在全局上下文中,所有节点都将访问存储的信息.
//
//   - 在每棵树上下文中,只有共享同一棵树的节点共享存储的信息.
//
//   - 在每个节点每个树上下文中, 黑板中存储的信息只能被写入数据的同一节点访问.
//
//     上下文是通过提供给这些方法的参数间接选择的,例如:
//
//     // getting/setting variable in global context
//     blackboard.set('testKey', 'value');
//     var value = blackboard.get('testKey');
//
//     // getting/setting variable in per tree context
//     blackboard.set('testKey', 'value', tree.id);
//     var value = blackboard.get('testKey', tree.id);
//
//     // getting/setting variable in per node per tree context
//     blackboard.set('testKey', 'value', tree.id, node.id);
//     var value = blackboard.get('testKey', tree.id, node.id);
//
// Note: 在内部, 黑板将这些内存存储在不同的对象中, 即"baseMemory"上的全局内存、
// "treeMemory" 上的每棵树以及在每棵树内存中动态创建的每棵树的每个节点(通过"treeMemory[id]"访问).
// 节点内存`).避免手动使用这些变量，而是使用"get"和"set".
//

// ------------------------TreeData-------------------------
type TreeData struct {
	NodeMemory     *Memory
	OpenNodes      []IBaseNode
	TraversalDepth int
	TraversalCycle int
}

func NewTreeData() *TreeData {
	return &TreeData{NewMemory(), make([]IBaseNode, 0), 0, 0}
}

// ------------------------Memory-------------------------
type Memory struct {
	m map[string]interface{}
}

func NewMemory() *Memory {
	return &Memory{make(map[string]interface{})}
}

// Get 返回Key对应的Value
func (m *Memory) Get(key string) interface{} {
	return m.m[key]
}

// Set 设置Key/Value
func (m *Memory) Set(key string, val interface{}) {
	m.m[key] = val
}

// Remove 删除对应的key
func (m *Memory) Remove(key string) {
	delete(m.m, key)
}

// ------------------------TreeMemory-------------------------
type TreeMemory struct {
	*Memory
	treeData   *TreeData
	nodeMemory map[string]*Memory
}

func NewTreeMemory() *TreeMemory {
	return &TreeMemory{NewMemory(), NewTreeData(), make(map[string]*Memory)}
}

// ------------------------Blackboard-------------------------
type Blackboard struct {
	baseMemory *Memory
	treeMemory map[string]*TreeMemory
}

func NewBlackboard() *Blackboard {
	p := &Blackboard{}
	p.Initialize()
	return p
}

func (bb *Blackboard) Initialize() {
	bb.baseMemory = NewMemory()
	bb.treeMemory = make(map[string]*TreeMemory)
}

// 检索树上下文内存的内部方法. 如果内存不存在, 则此方法创建它.
// treeScope 范围内树的 id
// 返回 The tree memory
func (bb *Blackboard) getTreeMemory(treeScope string) *TreeMemory {
	if _, ok := bb.treeMemory[treeScope]; !ok {
		bb.treeMemory[treeScope] = NewTreeMemory()
	}
	return bb.treeMemory[treeScope]
}

// getNodeMemory 给定树内存,检索节点上下文内存的内部方法.如果内存不存在, 则此方法创建内存.
// treeMemory 当前树内存
// nodeScope  作用域内节点的id
func (bb *Blackboard) getNodeMemory(treeMemory *TreeMemory, nodeScope string) *Memory {
	memory := treeMemory.nodeMemory
	if _, ok := memory[nodeScope]; !ok {
		memory[nodeScope] = NewMemory()
	}

	return memory[nodeScope]
}

// 检索上下文内存的内部方法. 如果提供了treeScope和nodeScope, 则此方法返回每个节点每个树的内存.
// 如果仅提供了treeScope, 则它返回每棵树的内存。如果没有提供参数, 则返回全局内存. 请注意, 如果
// 仅提供了nodeScope，此方法仍将返回全局内存.
// treeScope 树范围的ID.
// nodeScope 节点范围的id.
// 返回内存对象
func (bb *Blackboard) getMemory(treeScope, nodeScope string) *Memory {
	var memory = bb.baseMemory

	if len(treeScope) > 0 {
		treeMem := bb.getTreeMemory(treeScope)
		memory = treeMem.Memory
		if len(nodeScope) > 0 {
			memory = bb.getNodeMemory(treeMem, nodeScope)
		}
	}

	return memory
}

// 在黑板上存储一个值. 如果提供了treeScope和nodeScope, 此方法会将值保存到每个节点每个树内存中.
// 如果只提供了treeScope, 它会将值保存到每棵树的内存中. 如果没有提供参数，该方法会将值保存到全局
// 内存中.请注意, 如果仅提供了nodeScope(但没有提供treeScope),此方法仍会将值保存到全局内存中.
// key 要存储的关键字
// value 要存储的对像
// treeScope 访问树或节点内存时的树 ID.
// nodeScope 访问节点内存时的节点 ID.
func (bb *Blackboard) Set(key string, value interface{}, treeScope, nodeScope string) {
	var memory = bb.getMemory(treeScope, nodeScope)
	memory.Set(key, value)
}

func (bb *Blackboard) SetMem(key string, value interface{}) {
	var memory = bb.getMemory("", "")
	memory.Set(key, value)
}

func (bb *Blackboard) Remove(key string) {
	var memory = bb.getMemory("", "")
	memory.Remove(key)
}
func (bb *Blackboard) SetTree(key string, value interface{}, treeScope string) {
	var memory = bb.getMemory(treeScope, "")
	memory.Set(key, value)
}
func (bb *Blackboard) getTreeData(treeScope string) *TreeData {
	treeMem := bb.getTreeMemory(treeScope)
	return treeMem.treeData
}

// 检索黑板上的值.如果提供了treeScope和nodeScope,此方法将从每个节点每个树内存中检索值.
// 如果仅提供了treeScope,它将从每个树内存中检索值.如果没有提供参数,该方法将从全局内存中
// 检索.如果仅提供了nodeScope(但没有提供treeScope),此方法仍将尝试从全局内存中检索.
// key 要检索的关键字
// treeScope 访问树或节点内存时的树 ID.
// nodeScope 访问节点内存时的节点 ID.
// 返回 已存储或未定义的值
func (bb *Blackboard) Get(key, treeScope, nodeScope string) interface{} {
	memory := bb.getMemory(treeScope, nodeScope)
	return memory.Get(key)
}

func (bb *Blackboard) GetMem(key string) interface{} {
	memory := bb.getMemory("", "")
	return memory.Get(key)
}

func (bb *Blackboard) GetFloat64(key, treeScope, nodeScope string) float64 {
	v := bb.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return v.(float64)
}

func (bb *Blackboard) GetBool(key, treeScope, nodeScope string) bool {
	v := bb.Get(key, treeScope, nodeScope)
	if v == nil {
		return false
	}
	return v.(bool)
}

func (bb *Blackboard) GetInt(key, treeScope, nodeScope string) int {
	v := bb.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return v.(int)
}
func (bb *Blackboard) GetInt64(key, treeScope, nodeScope string) int64 {
	v := bb.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return v.(int64)
}

func (bb *Blackboard) GetUInt64(key, treeScope, nodeScope string) uint64 {
	v := bb.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return v.(uint64)
}

func (bb *Blackboard) GetInt64Safe(key, treeScope, nodeScope string) int64 {
	v := bb.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return ReadNumberToInt64(v)
}
func (bb *Blackboard) GetUInt64Safe(key, treeScope, nodeScope string) uint64 {
	v := bb.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return ReadNumberToUInt64(v)
}

func (bb *Blackboard) GetInt32(key, treeScope, nodeScope string) int32 {
	v := bb.Get(key, treeScope, nodeScope)
	if v == nil {
		return 0
	}
	return v.(int32)
}

func ReadNumberToInt64(v interface{}) int64 {
	var ret int64
	switch tvalue := v.(type) {
	case uint64:
		ret = int64(tvalue)
	default:
		panic(fmt.Sprintf("错误的类型转成Int64 %v:%+v", reflect.TypeOf(v), v))
	}

	return ret
}

func ReadNumberToUInt64(v interface{}) uint64 {
	var ret uint64
	switch tvalue := v.(type) {
	case int64:
		ret = uint64(tvalue)
	default:
		panic(fmt.Sprintf("错误的类型转成UInt64 %v:%+v", reflect.TypeOf(v), v))
	}
	return ret
}

//
//func ReadNumberToInt32(v interface{}) int32 {
//	var ret int32
//	switch tvalue := v.(type) {
//	case uint16, int16,uint32, int32,uint64,int64,uint16, int16,int:
//		ret = int32(tvalue)
//	default:
//		panic(fmt.Sprintf("错误的类型转成Int32 %v:%+v", reflect.TypeOf(v), v))
//	}
//	return ret
//}
//
//func ReadNumberToUInt32(v interface{}) uint32 {
//	var ret uint32
//	switch tvalue := v.(type) {
//	case uint16, int16,uint32, int32,uint64,int64,uint16, int16,int:
//		ret = uint32(tvalue)
//	default:
//		panic(fmt.Sprintf("错误的类型转成UInt32 %v:%+v", reflect.TypeOf(v), v))
//	}
//	return ret
//}
//
//
//func ReadNumberToInt(v interface{}) int {
//	var ret int
//	switch tvalue := v.(type) {
//	case uint16, int16,uint32, int32,uint64,int64,uint16, int16,int:
//		ret = int(tvalue)
//	default:
//		panic(fmt.Sprintf("int %v:%+v", reflect.TypeOf(v), v))
//	}
//	return ret
//}
