package hashring

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strconv"
)

var defaultHashFunc = func() HashFunc {
	hashFunc, err := NewHash(md5.New).Use(NewInt64PairHashKey)
	if err != nil {
		panic(fmt.Sprintf("failed to create defaultHashFunc: %s", err.Error()))
	}
	return hashFunc
}()

//HashKey hash key
type HashKey interface {
	Less(other HashKey) bool
}

//HashKeyOrder hash key order
type HashKeyOrder []HashKey

//排序函数
func (h HashKeyOrder) Len() int           { return len(h) }
func (h HashKeyOrder) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h HashKeyOrder) Less(i, j int) bool { return h[i].Less(h[j]) }

//HashFunc  hash 回调函数
type HashFunc func([]byte) HashKey

//Uint32HashKey hash 值
type Uint32HashKey uint32

//Less 排序方法（降序）
func (k Uint32HashKey) Less(other HashKey) bool {
	return k < other.(Uint32HashKey)
}

//HashRing hash 负载环
type HashRing struct {
	_ring       map[HashKey]string
	_sortedKeys []HashKey
	_nodes      []string
	_weights    map[string]int
	_hashFunc   HashFunc
}

//New 使用默认hash创建一个hash环
func New(nodes []string) *HashRing {
	return NewWithHash(nodes, defaultHashFunc)
}

//NewWithHash 自定义hash方法创建一个hash环
func NewWithHash(nodes []string, hashKey HashFunc) *HashRing {
	hashRing := &HashRing{
		_ring:       make(map[HashKey]string),
		_sortedKeys: make([]HashKey, 0),
		_nodes:      nodes,
		_weights:    make(map[string]int),
		_hashFunc:   hashKey,
	}
	hashRing.generateCircle()
	return hashRing
}

//NewWithWeights 自定义权重创建hash环
func NewWithWeights(weights map[string]int) *HashRing {
	return NewWithHashAndWeights(weights, defaultHashFunc)
}

//NewWithHashAndWeights 自定义权重及Hash方法，创建hash环
func NewWithHashAndWeights(weights map[string]int, hashFunc HashFunc) *HashRing {
	nodes := make([]string, 0, len(weights))
	for node := range weights {
		nodes = append(nodes, node)
	}
	hashRing := &HashRing{
		_ring:       make(map[HashKey]string),
		_sortedKeys: make([]HashKey, 0),
		_nodes:      nodes,
		_weights:    weights,
		_hashFunc:   hashFunc,
	}
	hashRing.generateCircle()
	return hashRing
}

//Size 返回节点大小
func (slf *HashRing) Size() int {
	return len(slf._nodes)
}

//UpdateWithWeights 更新权重
func (slf *HashRing) UpdateWithWeights(weights map[string]int) {
	nodesChgFlg := false
	if len(weights) != len(slf._weights) {
		nodesChgFlg = true
	} else {
		for node, newWeight := range weights {
			oldWeight, ok := slf._weights[node]
			if !ok || oldWeight != newWeight {
				nodesChgFlg = true
				break
			}
		}
	}

	if nodesChgFlg {
		newhring := NewWithHashAndWeights(weights, slf._hashFunc)
		slf._weights = newhring._weights
		slf._nodes = newhring._nodes
		slf._ring = newhring._ring
		slf._sortedKeys = newhring._sortedKeys
	}
}

func (slf *HashRing) generateCircle() {
	totalWeight := 0
	for _, node := range slf._nodes {
		if weight, ok := slf._weights[node]; ok {
			totalWeight += weight
		} else {
			totalWeight++
			slf._weights[node] = 1
		}
	}

	for _, node := range slf._nodes {
		weight := slf._weights[node]

		for j := 0; j < weight; j++ {
			nodeKey := node + "-" + strconv.FormatInt(int64(j), 10)
			key := slf._hashFunc([]byte(nodeKey))
			slf._ring[key] = node
			slf._sortedKeys = append(slf._sortedKeys, key)
		}
	}

	sort.Sort(HashKeyOrder(slf._sortedKeys))
}

//GetNode 返回一个节点
func (slf *HashRing) GetNode(stringKey string) (node string, ok bool) {
	pos, ok := slf.GetNodePos(stringKey)
	if !ok {
		return "", false
	}
	return slf._ring[slf._sortedKeys[pos]], true
}

//GetNodePos 翻译一个节点的位置
func (slf *HashRing) GetNodePos(stringKey string) (pos int, ok bool) {
	if len(slf._ring) == 0 {
		return 0, false
	}

	key := slf.GenKey(stringKey)

	nodes := slf._sortedKeys
	pos = sort.Search(len(nodes), func(i int) bool { return key.Less(nodes[i]) })

	if pos == len(nodes) {
		// Wrap the search, should return First node
		return 0, true
	}
	return pos, true
}

//GenKey 生成一个HaskKey
func (slf *HashRing) GenKey(key string) HashKey {
	return slf._hashFunc([]byte(key))
}

// GetNodes iterates over the hash ring and returns the nodes in the order
// which is determined by the key. GetNodes is thread safe if the hash
// which was used to configure the hash ring is thread safe.
func (slf *HashRing) GetNodes(stringKey string, size int) (nodes []string, ok bool) {
	pos, ok := slf.GetNodePos(stringKey)
	if !ok {
		return nil, false
	}

	if size > len(slf._nodes) {
		return nil, false
	}

	returnedValues := make(map[string]bool, size)
	//mergedSortedKeys := append(h.sortedKeys[pos:], h.sortedKeys[:pos]...)
	resultSlice := make([]string, 0, size)

	for i := pos; i < pos+len(slf._sortedKeys); i++ {
		key := slf._sortedKeys[i%len(slf._sortedKeys)]
		val := slf._ring[key]
		if !returnedValues[val] {
			returnedValues[val] = true
			resultSlice = append(resultSlice, val)
		}
		if len(returnedValues) == size {
			break
		}
	}

	return resultSlice, len(resultSlice) == size
}

//AddNode 增加一个节点
func (slf *HashRing) AddNode(node string) *HashRing {
	return slf.AddWeightedNode(node, 1)
}

//AddWeightedNode 增加一个权重节点
//返回Hash环
func (slf *HashRing) AddWeightedNode(node string, weight int) *HashRing {
	if weight <= 0 {
		return slf
	}

	if _, ok := slf._weights[node]; ok {
		return slf
	}

	nodes := make([]string, len(slf._nodes), len(slf._nodes)+1)
	copy(nodes, slf._nodes)
	nodes = append(nodes, node)

	weights := make(map[string]int)
	for eNode, eWeight := range slf._weights {
		weights[eNode] = eWeight
	}
	weights[node] = weight

	hashRing := &HashRing{
		_ring:       make(map[HashKey]string),
		_sortedKeys: make([]HashKey, 0),
		_nodes:      nodes,
		_weights:    weights,
		_hashFunc:   slf._hashFunc,
	}
	hashRing.generateCircle()
	return hashRing
}

//UpdateWeightedNode 更新一个节点权重
//返回Hash环
func (slf *HashRing) UpdateWeightedNode(node string, weight int) *HashRing {
	if weight <= 0 {
		return slf
	}

	/* node is not need to update for node is not existed or weight is not changed */
	if oldWeight, ok := slf._weights[node]; (!ok) || (ok && oldWeight == weight) {
		return slf
	}

	nodes := make([]string, len(slf._nodes))
	copy(nodes, slf._nodes)

	weights := make(map[string]int)
	for eNode, eWeight := range slf._weights {
		weights[eNode] = eWeight
	}
	weights[node] = weight

	hashRing := &HashRing{
		_ring:       make(map[HashKey]string),
		_sortedKeys: make([]HashKey, 0),
		_nodes:      nodes,
		_weights:    weights,
		_hashFunc:   slf._hashFunc,
	}
	hashRing.generateCircle()
	return hashRing
}

//RemoveNode 移除一个节点
func (slf *HashRing) RemoveNode(node string) *HashRing {
	/* if node isn't exist in hashring, don't refresh hashring */
	if _, ok := slf._weights[node]; !ok {
		return slf
	}

	nodes := make([]string, 0)
	for _, eNode := range slf._nodes {
		if eNode != node {
			nodes = append(nodes, eNode)
		}
	}

	weights := make(map[string]int)
	for eNode, eWeight := range slf._weights {
		if eNode != node {
			weights[eNode] = eWeight
		}
	}

	hashRing := &HashRing{
		_ring:       make(map[HashKey]string),
		_sortedKeys: make([]HashKey, 0),
		_nodes:      nodes,
		_weights:    weights,
		_hashFunc:   slf._hashFunc,
	}
	hashRing.generateCircle()
	return hashRing
}
