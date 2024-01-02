package skipset

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// Float32Set represents a set based on skip list in ascending order.
type Float32Set struct {
	header       *float32Node
	length       int64
	highestLevel int64 // highest level for now
}

type float32Node struct {
	value float32
	next  optionalArray // [level]*float32Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newFloat32Node(value float32, level int) *float32Node {
	node := &float32Node{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *float32Node) loadNext(i int) *float32Node {
	return (*float32Node)(n.next.load(i))
}

func (n *float32Node) storeNext(i int, node *float32Node) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *float32Node) atomicLoadNext(i int) *float32Node {
	return (*float32Node)(n.next.atomicLoad(i))
}

func (n *float32Node) atomicStoreNext(i int, node *float32Node) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *float32Node) lessthan(value float32) bool {
	return n.value < value
}

func (n *float32Node) equal(value float32) bool {
	return n.value == value
}

// NewFloat32 return an empty float32 skip set in ascending order.
func NewFloat32() *Float32Set {
	h := newFloat32Node(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Float32Set{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Float32Set) findNodeRemove(value float32, preds *[maxLevel]*float32Node, succs *[maxLevel]*float32Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Float32Set) findNodeAdd(value float32, preds *[maxLevel]*float32Node, succs *[maxLevel]*float32Node) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockFloat32(preds [maxLevel]*float32Node, highestLevel int) {
	var prevPred *float32Node
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Float32Set) Add(value float32) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*float32Node
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float32Node
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockFloat32(preds, highestLocked)
			continue
		}

		nn := newFloat32Node(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Float32Set) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Float32Set) Contains(value float32) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Float32Set) Remove(value float32) bool {
	var (
		nodeToRemove *float32Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float32Node
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *float32Node
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockFloat32(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockFloat32(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Float32Set) Range(f func(value float32) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Float32Set) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Float32SetDesc represents a set based on skip list in descending order.
type Float32SetDesc struct {
	header       *float32NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type float32NodeDesc struct {
	value float32
	next  optionalArray // [level]*float32NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newFloat32NodeDesc(value float32, level int) *float32NodeDesc {
	node := &float32NodeDesc{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *float32NodeDesc) loadNext(i int) *float32NodeDesc {
	return (*float32NodeDesc)(n.next.load(i))
}

func (n *float32NodeDesc) storeNext(i int, node *float32NodeDesc) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *float32NodeDesc) atomicLoadNext(i int) *float32NodeDesc {
	return (*float32NodeDesc)(n.next.atomicLoad(i))
}

func (n *float32NodeDesc) atomicStoreNext(i int, node *float32NodeDesc) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *float32NodeDesc) lessthan(value float32) bool {
	return n.value > value
}

func (n *float32NodeDesc) equal(value float32) bool {
	return n.value == value
}

// NewFloat32Desc return an empty float32 skip set in descending order.
func NewFloat32Desc() *Float32SetDesc {
	h := newFloat32NodeDesc(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Float32SetDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Float32SetDesc) findNodeRemove(value float32, preds *[maxLevel]*float32NodeDesc, succs *[maxLevel]*float32NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Float32SetDesc) findNodeAdd(value float32, preds *[maxLevel]*float32NodeDesc, succs *[maxLevel]*float32NodeDesc) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockFloat32Desc(preds [maxLevel]*float32NodeDesc, highestLevel int) {
	var prevPred *float32NodeDesc
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Float32SetDesc) Add(value float32) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*float32NodeDesc
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float32NodeDesc
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockFloat32Desc(preds, highestLocked)
			continue
		}

		nn := newFloat32NodeDesc(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Float32SetDesc) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Float32SetDesc) Contains(value float32) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Float32SetDesc) Remove(value float32) bool {
	var (
		nodeToRemove *float32NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float32NodeDesc
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *float32NodeDesc
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockFloat32Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockFloat32Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Float32SetDesc) Range(f func(value float32) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Float32SetDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Float64Set represents a set based on skip list in ascending order.
type Float64Set struct {
	header       *float64Node
	length       int64
	highestLevel int64 // highest level for now
}

type float64Node struct {
	value float64
	next  optionalArray // [level]*float64Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newFloat64Node(value float64, level int) *float64Node {
	node := &float64Node{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *float64Node) loadNext(i int) *float64Node {
	return (*float64Node)(n.next.load(i))
}

func (n *float64Node) storeNext(i int, node *float64Node) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *float64Node) atomicLoadNext(i int) *float64Node {
	return (*float64Node)(n.next.atomicLoad(i))
}

func (n *float64Node) atomicStoreNext(i int, node *float64Node) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *float64Node) lessthan(value float64) bool {
	return n.value < value
}

func (n *float64Node) equal(value float64) bool {
	return n.value == value
}

// NewFloat64 return an empty float64 skip set in ascending order.
func NewFloat64() *Float64Set {
	h := newFloat64Node(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Float64Set{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Float64Set) findNodeRemove(value float64, preds *[maxLevel]*float64Node, succs *[maxLevel]*float64Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Float64Set) findNodeAdd(value float64, preds *[maxLevel]*float64Node, succs *[maxLevel]*float64Node) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockFloat64(preds [maxLevel]*float64Node, highestLevel int) {
	var prevPred *float64Node
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Float64Set) Add(value float64) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*float64Node
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float64Node
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockFloat64(preds, highestLocked)
			continue
		}

		nn := newFloat64Node(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat64(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Float64Set) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Float64Set) Contains(value float64) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Float64Set) Remove(value float64) bool {
	var (
		nodeToRemove *float64Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float64Node
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *float64Node
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockFloat64(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockFloat64(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Float64Set) Range(f func(value float64) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Float64Set) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Float64SetDesc represents a set based on skip list in descending order.
type Float64SetDesc struct {
	header       *float64NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type float64NodeDesc struct {
	value float64
	next  optionalArray // [level]*float64NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newFloat64NodeDesc(value float64, level int) *float64NodeDesc {
	node := &float64NodeDesc{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *float64NodeDesc) loadNext(i int) *float64NodeDesc {
	return (*float64NodeDesc)(n.next.load(i))
}

func (n *float64NodeDesc) storeNext(i int, node *float64NodeDesc) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *float64NodeDesc) atomicLoadNext(i int) *float64NodeDesc {
	return (*float64NodeDesc)(n.next.atomicLoad(i))
}

func (n *float64NodeDesc) atomicStoreNext(i int, node *float64NodeDesc) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *float64NodeDesc) lessthan(value float64) bool {
	return n.value > value
}

func (n *float64NodeDesc) equal(value float64) bool {
	return n.value == value
}

// NewFloat64Desc return an empty float64 skip set in descending order.
func NewFloat64Desc() *Float64SetDesc {
	h := newFloat64NodeDesc(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Float64SetDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Float64SetDesc) findNodeRemove(value float64, preds *[maxLevel]*float64NodeDesc, succs *[maxLevel]*float64NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Float64SetDesc) findNodeAdd(value float64, preds *[maxLevel]*float64NodeDesc, succs *[maxLevel]*float64NodeDesc) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockFloat64Desc(preds [maxLevel]*float64NodeDesc, highestLevel int) {
	var prevPred *float64NodeDesc
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Float64SetDesc) Add(value float64) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*float64NodeDesc
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float64NodeDesc
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockFloat64Desc(preds, highestLocked)
			continue
		}

		nn := newFloat64NodeDesc(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat64Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Float64SetDesc) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Float64SetDesc) Contains(value float64) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Float64SetDesc) Remove(value float64) bool {
	var (
		nodeToRemove *float64NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float64NodeDesc
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *float64NodeDesc
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockFloat64Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockFloat64Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Float64SetDesc) Range(f func(value float64) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Float64SetDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Int32Set represents a set based on skip list in ascending order.
type Int32Set struct {
	header       *int32Node
	length       int64
	highestLevel int64 // highest level for now
}

type int32Node struct {
	value int32
	next  optionalArray // [level]*int32Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newInt32Node(value int32, level int) *int32Node {
	node := &int32Node{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *int32Node) loadNext(i int) *int32Node {
	return (*int32Node)(n.next.load(i))
}

func (n *int32Node) storeNext(i int, node *int32Node) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *int32Node) atomicLoadNext(i int) *int32Node {
	return (*int32Node)(n.next.atomicLoad(i))
}

func (n *int32Node) atomicStoreNext(i int, node *int32Node) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *int32Node) lessthan(value int32) bool {
	return n.value < value
}

func (n *int32Node) equal(value int32) bool {
	return n.value == value
}

// NewInt32 return an empty int32 skip set in ascending order.
func NewInt32() *Int32Set {
	h := newInt32Node(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Int32Set{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Int32Set) findNodeRemove(value int32, preds *[maxLevel]*int32Node, succs *[maxLevel]*int32Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Int32Set) findNodeAdd(value int32, preds *[maxLevel]*int32Node, succs *[maxLevel]*int32Node) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockInt32(preds [maxLevel]*int32Node, highestLevel int) {
	var prevPred *int32Node
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Int32Set) Add(value int32) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*int32Node
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int32Node
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockInt32(preds, highestLocked)
			continue
		}

		nn := newInt32Node(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Int32Set) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Int32Set) Contains(value int32) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Int32Set) Remove(value int32) bool {
	var (
		nodeToRemove *int32Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int32Node
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *int32Node
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockInt32(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockInt32(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Int32Set) Range(f func(value int32) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Int32Set) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Int32SetDesc represents a set based on skip list in descending order.
type Int32SetDesc struct {
	header       *int32NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type int32NodeDesc struct {
	value int32
	next  optionalArray // [level]*int32NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newInt32NodeDesc(value int32, level int) *int32NodeDesc {
	node := &int32NodeDesc{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *int32NodeDesc) loadNext(i int) *int32NodeDesc {
	return (*int32NodeDesc)(n.next.load(i))
}

func (n *int32NodeDesc) storeNext(i int, node *int32NodeDesc) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *int32NodeDesc) atomicLoadNext(i int) *int32NodeDesc {
	return (*int32NodeDesc)(n.next.atomicLoad(i))
}

func (n *int32NodeDesc) atomicStoreNext(i int, node *int32NodeDesc) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *int32NodeDesc) lessthan(value int32) bool {
	return n.value > value
}

func (n *int32NodeDesc) equal(value int32) bool {
	return n.value == value
}

// NewInt32Desc return an empty int32 skip set in descending order.
func NewInt32Desc() *Int32SetDesc {
	h := newInt32NodeDesc(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Int32SetDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Int32SetDesc) findNodeRemove(value int32, preds *[maxLevel]*int32NodeDesc, succs *[maxLevel]*int32NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Int32SetDesc) findNodeAdd(value int32, preds *[maxLevel]*int32NodeDesc, succs *[maxLevel]*int32NodeDesc) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockInt32Desc(preds [maxLevel]*int32NodeDesc, highestLevel int) {
	var prevPred *int32NodeDesc
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Int32SetDesc) Add(value int32) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*int32NodeDesc
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int32NodeDesc
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockInt32Desc(preds, highestLocked)
			continue
		}

		nn := newInt32NodeDesc(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Int32SetDesc) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Int32SetDesc) Contains(value int32) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Int32SetDesc) Remove(value int32) bool {
	var (
		nodeToRemove *int32NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int32NodeDesc
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *int32NodeDesc
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockInt32Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockInt32Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Int32SetDesc) Range(f func(value int32) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Int32SetDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Int16Set represents a set based on skip list in ascending order.
type Int16Set struct {
	header       *int16Node
	length       int64
	highestLevel int64 // highest level for now
}

type int16Node struct {
	value int16
	next  optionalArray // [level]*int16Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newInt16Node(value int16, level int) *int16Node {
	node := &int16Node{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *int16Node) loadNext(i int) *int16Node {
	return (*int16Node)(n.next.load(i))
}

func (n *int16Node) storeNext(i int, node *int16Node) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *int16Node) atomicLoadNext(i int) *int16Node {
	return (*int16Node)(n.next.atomicLoad(i))
}

func (n *int16Node) atomicStoreNext(i int, node *int16Node) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *int16Node) lessthan(value int16) bool {
	return n.value < value
}

func (n *int16Node) equal(value int16) bool {
	return n.value == value
}

// NewInt16 return an empty int16 skip set in ascending order.
func NewInt16() *Int16Set {
	h := newInt16Node(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Int16Set{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Int16Set) findNodeRemove(value int16, preds *[maxLevel]*int16Node, succs *[maxLevel]*int16Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Int16Set) findNodeAdd(value int16, preds *[maxLevel]*int16Node, succs *[maxLevel]*int16Node) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockInt16(preds [maxLevel]*int16Node, highestLevel int) {
	var prevPred *int16Node
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Int16Set) Add(value int16) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*int16Node
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int16Node
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockInt16(preds, highestLocked)
			continue
		}

		nn := newInt16Node(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt16(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Int16Set) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Int16Set) Contains(value int16) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Int16Set) Remove(value int16) bool {
	var (
		nodeToRemove *int16Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int16Node
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *int16Node
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockInt16(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockInt16(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Int16Set) Range(f func(value int16) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Int16Set) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Int16SetDesc represents a set based on skip list in descending order.
type Int16SetDesc struct {
	header       *int16NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type int16NodeDesc struct {
	value int16
	next  optionalArray // [level]*int16NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newInt16NodeDesc(value int16, level int) *int16NodeDesc {
	node := &int16NodeDesc{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *int16NodeDesc) loadNext(i int) *int16NodeDesc {
	return (*int16NodeDesc)(n.next.load(i))
}

func (n *int16NodeDesc) storeNext(i int, node *int16NodeDesc) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *int16NodeDesc) atomicLoadNext(i int) *int16NodeDesc {
	return (*int16NodeDesc)(n.next.atomicLoad(i))
}

func (n *int16NodeDesc) atomicStoreNext(i int, node *int16NodeDesc) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *int16NodeDesc) lessthan(value int16) bool {
	return n.value > value
}

func (n *int16NodeDesc) equal(value int16) bool {
	return n.value == value
}

// NewInt16Desc return an empty int16 skip set in descending order.
func NewInt16Desc() *Int16SetDesc {
	h := newInt16NodeDesc(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Int16SetDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Int16SetDesc) findNodeRemove(value int16, preds *[maxLevel]*int16NodeDesc, succs *[maxLevel]*int16NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Int16SetDesc) findNodeAdd(value int16, preds *[maxLevel]*int16NodeDesc, succs *[maxLevel]*int16NodeDesc) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockInt16Desc(preds [maxLevel]*int16NodeDesc, highestLevel int) {
	var prevPred *int16NodeDesc
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Int16SetDesc) Add(value int16) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*int16NodeDesc
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int16NodeDesc
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockInt16Desc(preds, highestLocked)
			continue
		}

		nn := newInt16NodeDesc(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt16Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Int16SetDesc) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Int16SetDesc) Contains(value int16) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Int16SetDesc) Remove(value int16) bool {
	var (
		nodeToRemove *int16NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int16NodeDesc
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *int16NodeDesc
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockInt16Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockInt16Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Int16SetDesc) Range(f func(value int16) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Int16SetDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// IntSet represents a set based on skip list in ascending order.
type IntSet struct {
	header       *intNode
	length       int64
	highestLevel int64 // highest level for now
}

type intNode struct {
	value int
	next  optionalArray // [level]*intNode
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newIntNode(value int, level int) *intNode {
	node := &intNode{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *intNode) loadNext(i int) *intNode {
	return (*intNode)(n.next.load(i))
}

func (n *intNode) storeNext(i int, node *intNode) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *intNode) atomicLoadNext(i int) *intNode {
	return (*intNode)(n.next.atomicLoad(i))
}

func (n *intNode) atomicStoreNext(i int, node *intNode) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *intNode) lessthan(value int) bool {
	return n.value < value
}

func (n *intNode) equal(value int) bool {
	return n.value == value
}

// NewInt return an empty int skip set in ascending order.
func NewInt() *IntSet {
	h := newIntNode(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &IntSet{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *IntSet) findNodeRemove(value int, preds *[maxLevel]*intNode, succs *[maxLevel]*intNode) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *IntSet) findNodeAdd(value int, preds *[maxLevel]*intNode, succs *[maxLevel]*intNode) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockInt(preds [maxLevel]*intNode, highestLevel int) {
	var prevPred *intNode
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *IntSet) Add(value int) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*intNode
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *intNode
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockInt(preds, highestLocked)
			continue
		}

		nn := newIntNode(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *IntSet) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *IntSet) Contains(value int) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *IntSet) Remove(value int) bool {
	var (
		nodeToRemove *intNode
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*intNode
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *intNode
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockInt(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockInt(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *IntSet) Range(f func(value int) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *IntSet) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// IntSetDesc represents a set based on skip list in descending order.
type IntSetDesc struct {
	header       *intNodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type intNodeDesc struct {
	value int
	next  optionalArray // [level]*intNodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newIntNodeDesc(value int, level int) *intNodeDesc {
	node := &intNodeDesc{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *intNodeDesc) loadNext(i int) *intNodeDesc {
	return (*intNodeDesc)(n.next.load(i))
}

func (n *intNodeDesc) storeNext(i int, node *intNodeDesc) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *intNodeDesc) atomicLoadNext(i int) *intNodeDesc {
	return (*intNodeDesc)(n.next.atomicLoad(i))
}

func (n *intNodeDesc) atomicStoreNext(i int, node *intNodeDesc) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *intNodeDesc) lessthan(value int) bool {
	return n.value > value
}

func (n *intNodeDesc) equal(value int) bool {
	return n.value == value
}

// NewIntDesc return an empty int skip set in descending order.
func NewIntDesc() *IntSetDesc {
	h := newIntNodeDesc(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &IntSetDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *IntSetDesc) findNodeRemove(value int, preds *[maxLevel]*intNodeDesc, succs *[maxLevel]*intNodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *IntSetDesc) findNodeAdd(value int, preds *[maxLevel]*intNodeDesc, succs *[maxLevel]*intNodeDesc) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockIntDesc(preds [maxLevel]*intNodeDesc, highestLevel int) {
	var prevPred *intNodeDesc
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *IntSetDesc) Add(value int) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*intNodeDesc
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *intNodeDesc
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockIntDesc(preds, highestLocked)
			continue
		}

		nn := newIntNodeDesc(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockIntDesc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *IntSetDesc) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *IntSetDesc) Contains(value int) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *IntSetDesc) Remove(value int) bool {
	var (
		nodeToRemove *intNodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*intNodeDesc
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *intNodeDesc
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockIntDesc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockIntDesc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *IntSetDesc) Range(f func(value int) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *IntSetDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint64Set represents a set based on skip list in ascending order.
type Uint64Set struct {
	header       *uint64Node
	length       int64
	highestLevel int64 // highest level for now
}

type uint64Node struct {
	value uint64
	next  optionalArray // [level]*uint64Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUuint64Node(value uint64, level int) *uint64Node {
	node := &uint64Node{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint64Node) loadNext(i int) *uint64Node {
	return (*uint64Node)(n.next.load(i))
}

func (n *uint64Node) storeNext(i int, node *uint64Node) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *uint64Node) atomicLoadNext(i int) *uint64Node {
	return (*uint64Node)(n.next.atomicLoad(i))
}

func (n *uint64Node) atomicStoreNext(i int, node *uint64Node) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *uint64Node) lessthan(value uint64) bool {
	return n.value < value
}

func (n *uint64Node) equal(value uint64) bool {
	return n.value == value
}

// NewUint64 return an empty uint64 skip set in ascending order.
func NewUint64() *Uint64Set {
	h := newUuint64Node(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint64Set{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint64Set) findNodeRemove(value uint64, preds *[maxLevel]*uint64Node, succs *[maxLevel]*uint64Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint64Set) findNodeAdd(value uint64, preds *[maxLevel]*uint64Node, succs *[maxLevel]*uint64Node) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockUint64(preds [maxLevel]*uint64Node, highestLevel int) {
	var prevPred *uint64Node
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Uint64Set) Add(value uint64) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint64Node
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint64Node
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockUint64(preds, highestLocked)
			continue
		}

		nn := newUuint64Node(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint64(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Uint64Set) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Uint64Set) Contains(value uint64) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Uint64Set) Remove(value uint64) bool {
	var (
		nodeToRemove *uint64Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint64Node
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *uint64Node
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockUint64(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockUint64(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Uint64Set) Range(f func(value uint64) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Uint64Set) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint64SetDesc represents a set based on skip list in descending order.
type Uint64SetDesc struct {
	header       *uint64NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type uint64NodeDesc struct {
	value uint64
	next  optionalArray // [level]*uint64NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUuint64NodeDescDesc(value uint64, level int) *uint64NodeDesc {
	node := &uint64NodeDesc{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint64NodeDesc) loadNext(i int) *uint64NodeDesc {
	return (*uint64NodeDesc)(n.next.load(i))
}

func (n *uint64NodeDesc) storeNext(i int, node *uint64NodeDesc) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *uint64NodeDesc) atomicLoadNext(i int) *uint64NodeDesc {
	return (*uint64NodeDesc)(n.next.atomicLoad(i))
}

func (n *uint64NodeDesc) atomicStoreNext(i int, node *uint64NodeDesc) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *uint64NodeDesc) lessthan(value uint64) bool {
	return n.value > value
}

func (n *uint64NodeDesc) equal(value uint64) bool {
	return n.value == value
}

// NewUint64Desc return an empty uint64 skip set in descending order.
func NewUint64Desc() *Uint64SetDesc {
	h := newUuint64NodeDescDesc(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint64SetDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint64SetDesc) findNodeRemove(value uint64, preds *[maxLevel]*uint64NodeDesc, succs *[maxLevel]*uint64NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint64SetDesc) findNodeAdd(value uint64, preds *[maxLevel]*uint64NodeDesc, succs *[maxLevel]*uint64NodeDesc) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockUint64Desc(preds [maxLevel]*uint64NodeDesc, highestLevel int) {
	var prevPred *uint64NodeDesc
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Uint64SetDesc) Add(value uint64) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint64NodeDesc
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint64NodeDesc
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockUint64Desc(preds, highestLocked)
			continue
		}

		nn := newUuint64NodeDescDesc(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint64Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Uint64SetDesc) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Uint64SetDesc) Contains(value uint64) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Uint64SetDesc) Remove(value uint64) bool {
	var (
		nodeToRemove *uint64NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint64NodeDesc
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *uint64NodeDesc
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockUint64Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockUint64Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Uint64SetDesc) Range(f func(value uint64) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Uint64SetDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint32Set represents a set based on skip list in ascending order.
type Uint32Set struct {
	header       *uint32Node
	length       int64
	highestLevel int64 // highest level for now
}

type uint32Node struct {
	value uint32
	next  optionalArray // [level]*uint32Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUint32Node(value uint32, level int) *uint32Node {
	node := &uint32Node{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint32Node) loadNext(i int) *uint32Node {
	return (*uint32Node)(n.next.load(i))
}

func (n *uint32Node) storeNext(i int, node *uint32Node) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *uint32Node) atomicLoadNext(i int) *uint32Node {
	return (*uint32Node)(n.next.atomicLoad(i))
}

func (n *uint32Node) atomicStoreNext(i int, node *uint32Node) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *uint32Node) lessthan(value uint32) bool {
	return n.value < value
}

func (n *uint32Node) equal(value uint32) bool {
	return n.value == value
}

// NewUint32 return an empty uint32 skip set in ascending order.
func NewUint32() *Uint32Set {
	h := newUint32Node(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint32Set{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint32Set) findNodeRemove(value uint32, preds *[maxLevel]*uint32Node, succs *[maxLevel]*uint32Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint32Set) findNodeAdd(value uint32, preds *[maxLevel]*uint32Node, succs *[maxLevel]*uint32Node) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockUint32(preds [maxLevel]*uint32Node, highestLevel int) {
	var prevPred *uint32Node
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Uint32Set) Add(value uint32) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint32Node
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint32Node
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockUint32(preds, highestLocked)
			continue
		}

		nn := newUint32Node(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Uint32Set) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Uint32Set) Contains(value uint32) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Uint32Set) Remove(value uint32) bool {
	var (
		nodeToRemove *uint32Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint32Node
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *uint32Node
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockUint32(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockUint32(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Uint32Set) Range(f func(value uint32) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Uint32Set) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint32SetDesc represents a set based on skip list in descending order.
type Uint32SetDesc struct {
	header       *uint32NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type uint32NodeDesc struct {
	value uint32
	next  optionalArray // [level]*uint32NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUint32NodeDesc(value uint32, level int) *uint32NodeDesc {
	node := &uint32NodeDesc{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint32NodeDesc) loadNext(i int) *uint32NodeDesc {
	return (*uint32NodeDesc)(n.next.load(i))
}

func (n *uint32NodeDesc) storeNext(i int, node *uint32NodeDesc) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *uint32NodeDesc) atomicLoadNext(i int) *uint32NodeDesc {
	return (*uint32NodeDesc)(n.next.atomicLoad(i))
}

func (n *uint32NodeDesc) atomicStoreNext(i int, node *uint32NodeDesc) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *uint32NodeDesc) lessthan(value uint32) bool {
	return n.value > value
}

func (n *uint32NodeDesc) equal(value uint32) bool {
	return n.value == value
}

// NewUint32Desc return an empty uint32 skip set in descending order.
func NewUint32Desc() *Uint32SetDesc {
	h := newUint32NodeDesc(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint32SetDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint32SetDesc) findNodeRemove(value uint32, preds *[maxLevel]*uint32NodeDesc, succs *[maxLevel]*uint32NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint32SetDesc) findNodeAdd(value uint32, preds *[maxLevel]*uint32NodeDesc, succs *[maxLevel]*uint32NodeDesc) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockUint32Desc(preds [maxLevel]*uint32NodeDesc, highestLevel int) {
	var prevPred *uint32NodeDesc
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Uint32SetDesc) Add(value uint32) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint32NodeDesc
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint32NodeDesc
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockUint32Desc(preds, highestLocked)
			continue
		}

		nn := newUint32NodeDesc(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Uint32SetDesc) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Uint32SetDesc) Contains(value uint32) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Uint32SetDesc) Remove(value uint32) bool {
	var (
		nodeToRemove *uint32NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint32NodeDesc
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *uint32NodeDesc
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockUint32Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockUint32Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Uint32SetDesc) Range(f func(value uint32) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Uint32SetDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint16Set represents a set based on skip list in ascending order.
type Uint16Set struct {
	header       *uint16Node
	length       int64
	highestLevel int64 // highest level for now
}

type uint16Node struct {
	value uint16
	next  optionalArray // [level]*uint16Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUint16Node(value uint16, level int) *uint16Node {
	node := &uint16Node{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint16Node) loadNext(i int) *uint16Node {
	return (*uint16Node)(n.next.load(i))
}

func (n *uint16Node) storeNext(i int, node *uint16Node) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *uint16Node) atomicLoadNext(i int) *uint16Node {
	return (*uint16Node)(n.next.atomicLoad(i))
}

func (n *uint16Node) atomicStoreNext(i int, node *uint16Node) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *uint16Node) lessthan(value uint16) bool {
	return n.value < value
}

func (n *uint16Node) equal(value uint16) bool {
	return n.value == value
}

// NewUint16 return an empty uint16 skip set in ascending order.
func NewUint16() *Uint16Set {
	h := newUint16Node(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint16Set{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint16Set) findNodeRemove(value uint16, preds *[maxLevel]*uint16Node, succs *[maxLevel]*uint16Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint16Set) findNodeAdd(value uint16, preds *[maxLevel]*uint16Node, succs *[maxLevel]*uint16Node) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockUint16(preds [maxLevel]*uint16Node, highestLevel int) {
	var prevPred *uint16Node
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Uint16Set) Add(value uint16) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint16Node
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint16Node
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockUint16(preds, highestLocked)
			continue
		}

		nn := newUint16Node(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint16(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Uint16Set) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Uint16Set) Contains(value uint16) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Uint16Set) Remove(value uint16) bool {
	var (
		nodeToRemove *uint16Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint16Node
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *uint16Node
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockUint16(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockUint16(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Uint16Set) Range(f func(value uint16) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Uint16Set) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint16SetDesc represents a set based on skip list in descending order.
type Uint16SetDesc struct {
	header       *uint16NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type uint16NodeDesc struct {
	value uint16
	next  optionalArray // [level]*uint16NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUint16NodeDesc(value uint16, level int) *uint16NodeDesc {
	node := &uint16NodeDesc{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint16NodeDesc) loadNext(i int) *uint16NodeDesc {
	return (*uint16NodeDesc)(n.next.load(i))
}

func (n *uint16NodeDesc) storeNext(i int, node *uint16NodeDesc) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *uint16NodeDesc) atomicLoadNext(i int) *uint16NodeDesc {
	return (*uint16NodeDesc)(n.next.atomicLoad(i))
}

func (n *uint16NodeDesc) atomicStoreNext(i int, node *uint16NodeDesc) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *uint16NodeDesc) lessthan(value uint16) bool {
	return n.value > value
}

func (n *uint16NodeDesc) equal(value uint16) bool {
	return n.value == value
}

// NewUint16Desc return an empty uint16 skip set in descending order.
func NewUint16Desc() *Uint16SetDesc {
	h := newUint16NodeDesc(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint16SetDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint16SetDesc) findNodeRemove(value uint16, preds *[maxLevel]*uint16NodeDesc, succs *[maxLevel]*uint16NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *Uint16SetDesc) findNodeAdd(value uint16, preds *[maxLevel]*uint16NodeDesc, succs *[maxLevel]*uint16NodeDesc) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockUint16Desc(preds [maxLevel]*uint16NodeDesc, highestLevel int) {
	var prevPred *uint16NodeDesc
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *Uint16SetDesc) Add(value uint16) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint16NodeDesc
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint16NodeDesc
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockUint16Desc(preds, highestLocked)
			continue
		}

		nn := newUint16NodeDesc(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint16Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *Uint16SetDesc) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *Uint16SetDesc) Contains(value uint16) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *Uint16SetDesc) Remove(value uint16) bool {
	var (
		nodeToRemove *uint16NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint16NodeDesc
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *uint16NodeDesc
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockUint16Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockUint16Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *Uint16SetDesc) Range(f func(value uint16) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *Uint16SetDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// UintSet represents a set based on skip list in ascending order.
type UintSet struct {
	header       *uintNode
	length       int64
	highestLevel int64 // highest level for now
}

type uintNode struct {
	value uint
	next  optionalArray // [level]*uintNode
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUintNode(value uint, level int) *uintNode {
	node := &uintNode{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uintNode) loadNext(i int) *uintNode {
	return (*uintNode)(n.next.load(i))
}

func (n *uintNode) storeNext(i int, node *uintNode) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *uintNode) atomicLoadNext(i int) *uintNode {
	return (*uintNode)(n.next.atomicLoad(i))
}

func (n *uintNode) atomicStoreNext(i int, node *uintNode) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *uintNode) lessthan(value uint) bool {
	return n.value < value
}

func (n *uintNode) equal(value uint) bool {
	return n.value == value
}

// NewUint return an empty uint skip set in ascending order.
func NewUint() *UintSet {
	h := newUintNode(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &UintSet{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *UintSet) findNodeRemove(value uint, preds *[maxLevel]*uintNode, succs *[maxLevel]*uintNode) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *UintSet) findNodeAdd(value uint, preds *[maxLevel]*uintNode, succs *[maxLevel]*uintNode) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockUint(preds [maxLevel]*uintNode, highestLevel int) {
	var prevPred *uintNode
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *UintSet) Add(value uint) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uintNode
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uintNode
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockUint(preds, highestLocked)
			continue
		}

		nn := newUintNode(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *UintSet) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *UintSet) Contains(value uint) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *UintSet) Remove(value uint) bool {
	var (
		nodeToRemove *uintNode
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uintNode
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *uintNode
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockUint(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockUint(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *UintSet) Range(f func(value uint) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *UintSet) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// UintSetDesc represents a set based on skip list in descending order.
type UintSetDesc struct {
	header       *uintNodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type uintNodeDesc struct {
	value uint
	next  optionalArray // [level]*uintNodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUintNodeDesc(value uint, level int) *uintNodeDesc {
	node := &uintNodeDesc{
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uintNodeDesc) loadNext(i int) *uintNodeDesc {
	return (*uintNodeDesc)(n.next.load(i))
}

func (n *uintNodeDesc) storeNext(i int, node *uintNodeDesc) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *uintNodeDesc) atomicLoadNext(i int) *uintNodeDesc {
	return (*uintNodeDesc)(n.next.atomicLoad(i))
}

func (n *uintNodeDesc) atomicStoreNext(i int, node *uintNodeDesc) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

func (n *uintNodeDesc) lessthan(value uint) bool {
	return n.value > value
}

func (n *uintNodeDesc) equal(value uint) bool {
	return n.value == value
}

// NewUintDesc return an empty uint skip set in descending order.
func NewUintDesc() *UintSetDesc {
	h := newUintNodeDesc(0, maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &UintSetDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *UintSetDesc) findNodeRemove(value uint, preds *[maxLevel]*uintNodeDesc, succs *[maxLevel]*uintNodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(value) {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *UintSetDesc) findNodeAdd(value uint, preds *[maxLevel]*uintNodeDesc, succs *[maxLevel]*uintNodeDesc) int {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(value) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.equal(value) {
			return i
		}
	}
	return -1
}

func unlockUintDesc(preds [maxLevel]*uintNodeDesc, highestLevel int) {
	var prevPred *uintNodeDesc
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *UintSetDesc) Add(value uint) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uintNodeDesc
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uintNodeDesc
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockUintDesc(preds, highestLocked)
			continue
		}

		nn := newUintNodeDesc(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUintDesc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *UintSetDesc) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *UintSetDesc) Contains(value uint) bool {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(value) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.equal(value) {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *UintSetDesc) Remove(value uint) bool {
	var (
		nodeToRemove *uintNodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uintNodeDesc
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *uintNodeDesc
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockUintDesc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockUintDesc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *UintSetDesc) Range(f func(value uint) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *UintSetDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// StringSet represents a set based on skip list.
type StringSet struct {
	header       *stringNode
	length       int64
	highestLevel int64 // highest level for now
}

type stringNode struct {
	value string
	score uint64
	next  optionalArray // [level]*stringNode
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newStringNode(value string, level int) *stringNode {
	node := &stringNode{
		score: hash(value),
		value: value,
		level: uint32(level),
	}
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *stringNode) loadNext(i int) *stringNode {
	return (*stringNode)(n.next.load(i))
}

func (n *stringNode) storeNext(i int, node *stringNode) {
	n.next.store(i, unsafe.Pointer(node))
}

func (n *stringNode) atomicLoadNext(i int) *stringNode {
	return (*stringNode)(n.next.atomicLoad(i))
}

func (n *stringNode) atomicStoreNext(i int, node *stringNode) {
	n.next.atomicStore(i, unsafe.Pointer(node))
}

// NewString return an empty string skip set.
func NewString() *StringSet {
	h := newStringNode("", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &StringSet{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNodeRemove takes a value and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *StringSet) findNodeRemove(value string, preds *[maxLevel]*stringNode, succs *[maxLevel]*stringNode) int {
	score := hash(value)
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.cmp(score, value) < 0 {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if lFound == -1 && succ != nil && succ.cmp(score, value) == 0 {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a value and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > value >= succs[i].
func (s *StringSet) findNodeAdd(value string, preds *[maxLevel]*stringNode, succs *[maxLevel]*stringNode) int {
	score := hash(value)
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.cmp(score, value) < 0 {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the value already in the skip list.
		if succ != nil && succ.cmp(score, value) == 0 {
			return i
		}
	}
	return -1
}

func unlockString(preds [maxLevel]*stringNode, highestLevel int) {
	var prevPred *stringNode
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the value into skip set, return true if this process insert the value into skip set,
// return false if this process can't insert this value, because another process has insert the same value.
//
// If the value is in the skip set but not fully linked, this process will wait until it is.
func (s *StringSet) Add(value string) bool {
	level := s.randomlevel()
	var preds, succs [maxLevel]*stringNode
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the value is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}
		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *stringNode
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && (succ == nil || !succ.flags.Get(marked)) && pred.loadNext(layer) == succ
		}
		if !valid {
			unlockString(preds, highestLocked)
			continue
		}

		nn := newStringNode(value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockString(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

func (s *StringSet) randomlevel() int {
	// Generate random level.
	level := randomLevel()
	// Update highest level if possible.
	for {
		hl := atomic.LoadInt64(&s.highestLevel)
		if int64(level) <= hl {
			break
		}
		if atomic.CompareAndSwapInt64(&s.highestLevel, hl, int64(level)) {
			break
		}
	}
	return level
}

// Contains check if the value is in the skip set.
func (s *StringSet) Contains(value string) bool {
	score := hash(value)
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.cmp(score, value) < 0 {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the value already in the skip list.
		if nex != nil && nex.cmp(score, value) == 0 {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *StringSet) Remove(value string) bool {
	var (
		nodeToRemove *stringNode
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*stringNode
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *stringNode
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ
			}
			if !valid {
				unlockString(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].atomicStoreNext(i, nodeToRemove.loadNext(i))
			}
			nodeToRemove.mu.Unlock()
			unlockString(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each value present in the skip set.
// If f returns false, range stops the iteration.
func (s *StringSet) Range(f func(value string) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skip set.
func (s *StringSet) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Return 1 if n is bigger, 0 if equal, else -1.
func (n *stringNode) cmp(score uint64, value string) int {
	if n.score > score {
		return 1
	} else if n.score == score {
		return cmpstring(n.value, value)
	}
	return -1
}
