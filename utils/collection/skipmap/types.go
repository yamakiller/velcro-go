package skipmap

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// Float32Map represents a map based on skip list in ascending order.
type Float32Map struct {
	header       *float32Node
	length       int64
	highestLevel int64 // highest level for now
}

type float32Node struct {
	key   float32
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*float32Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newFloat32Node(key float32, value interface{}, level int) *float32Node {
	node := &float32Node{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *float32Node) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *float32Node) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *float32Node) lessthan(key float32) bool {
	return n.key < key
}

func (n *float32Node) equal(key float32) bool {
	return n.key == key
}

// NewFloat32 return an empty float32 skipmap.
func NewFloat32() *Float32Map {
	h := newFloat32Node(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Float32Map{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Float32Map) findNode(key float32, preds *[maxLevel]*float32Node, succs *[maxLevel]*float32Node) *float32Node {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Float32Map) findNodeDelete(key float32, preds *[maxLevel]*float32Node, succs *[maxLevel]*float32Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Float32Map) Store(key float32, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*float32Node
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newFloat32Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Float32Map) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Float32Map) Load(key float32) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Float32Map) LoadAndDelete(key float32) (value interface{}, loaded bool) {
	var (
		nodeToDelete *float32Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float32Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockFloat32(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Float32Map) LoadOrStore(key float32, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*float32Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float32Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newFloat32Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Float32Map) LoadOrStoreLazy(key float32, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*float32Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float32Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockFloat32(preds, highestLocked)
			continue
		}
		value := f()
		nn := newFloat32Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Float32Map) Delete(key float32) bool {
	var (
		nodeToDelete *float32Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float32Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockFloat32(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockFloat32(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Float32Map) Range(f func(key float32, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Float32Map) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Float32MapDesc represents a map based on skip list in descending order.
type Float32MapDesc struct {
	header       *float32NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type float32NodeDesc struct {
	key   float32
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*float32NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newFloat32NodeDesc(key float32, value interface{}, level int) *float32NodeDesc {
	node := &float32NodeDesc{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *float32NodeDesc) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *float32NodeDesc) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *float32NodeDesc) lessthan(key float32) bool {
	return n.key > key
}

func (n *float32NodeDesc) equal(key float32) bool {
	return n.key == key
}

// NewFloat32Desc return an empty float32 skipmap.
func NewFloat32Desc() *Float32MapDesc {
	h := newFloat32NodeDesc(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Float32MapDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Float32MapDesc) findNode(key float32, preds *[maxLevel]*float32NodeDesc, succs *[maxLevel]*float32NodeDesc) *float32NodeDesc {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Float32MapDesc) findNodeDelete(key float32, preds *[maxLevel]*float32NodeDesc, succs *[maxLevel]*float32NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Float32MapDesc) Store(key float32, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*float32NodeDesc
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newFloat32NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Float32MapDesc) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Float32MapDesc) Load(key float32) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Float32MapDesc) LoadAndDelete(key float32) (value interface{}, loaded bool) {
	var (
		nodeToDelete *float32NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float32NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockFloat32Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Float32MapDesc) LoadOrStore(key float32, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*float32NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float32NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newFloat32NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Float32MapDesc) LoadOrStoreLazy(key float32, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*float32NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float32NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockFloat32Desc(preds, highestLocked)
			continue
		}
		value := f()
		nn := newFloat32NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Float32MapDesc) Delete(key float32) bool {
	var (
		nodeToDelete *float32NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float32NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockFloat32Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockFloat32Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Float32MapDesc) Range(f func(key float32, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Float32MapDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Float64Map represents a map based on skip list in ascending order.
type Float64Map struct {
	header       *float64Node
	length       int64
	highestLevel int64 // highest level for now
}

type float64Node struct {
	key   float64
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*float64Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newFloat64Node(key float64, value interface{}, level int) *float64Node {
	node := &float64Node{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *float64Node) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *float64Node) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *float64Node) lessthan(key float64) bool {
	return n.key < key
}

func (n *float64Node) equal(key float64) bool {
	return n.key == key
}

// NewFloat64 return an empty float64 skipmap.
func NewFloat64() *Float64Map {
	h := newFloat64Node(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Float64Map{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Float64Map) findNode(key float64, preds *[maxLevel]*float64Node, succs *[maxLevel]*float64Node) *float64Node {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Float64Map) findNodeDelete(key float64, preds *[maxLevel]*float64Node, succs *[maxLevel]*float64Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Float64Map) Store(key float64, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*float64Node
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newFloat64Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat64(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Float64Map) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Float64Map) Load(key float64) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Float64Map) LoadAndDelete(key float64) (value interface{}, loaded bool) {
	var (
		nodeToDelete *float64Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float64Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockFloat64(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Float64Map) LoadOrStore(key float64, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*float64Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float64Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newFloat64Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat64(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Float64Map) LoadOrStoreLazy(key float64, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*float64Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float64Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockFloat64(preds, highestLocked)
			continue
		}
		value := f()
		nn := newFloat64Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat64(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Float64Map) Delete(key float64) bool {
	var (
		nodeToDelete *float64Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float64Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockFloat64(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockFloat64(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Float64Map) Range(f func(key float64, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Float64Map) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Float64MapDesc represents a map based on skip list in descending order.
type Float64MapDesc struct {
	header       *float64NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type float64NodeDesc struct {
	key   float64
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*float64NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newFloat64NodeDesc(key float64, value interface{}, level int) *float64NodeDesc {
	node := &float64NodeDesc{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *float64NodeDesc) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *float64NodeDesc) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *float64NodeDesc) lessthan(key float64) bool {
	return n.key > key
}

func (n *float64NodeDesc) equal(key float64) bool {
	return n.key == key
}

// NewFloat64Desc return an empty float64 skipmap.
func NewFloat64Desc() *Float64MapDesc {
	h := newFloat64NodeDesc(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Float64MapDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Float64MapDesc) findNode(key float64, preds *[maxLevel]*float64NodeDesc, succs *[maxLevel]*float64NodeDesc) *float64NodeDesc {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Float64MapDesc) findNodeDelete(key float64, preds *[maxLevel]*float64NodeDesc, succs *[maxLevel]*float64NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Float64MapDesc) Store(key float64, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*float64NodeDesc
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newFloat64NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat64Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Float64MapDesc) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Float64MapDesc) Load(key float64) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Float64MapDesc) LoadAndDelete(key float64) (value interface{}, loaded bool) {
	var (
		nodeToDelete *float64NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float64NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockFloat64Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Float64MapDesc) LoadOrStore(key float64, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*float64NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float64NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newFloat64NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat64Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Float64MapDesc) LoadOrStoreLazy(key float64, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*float64NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *float64NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockFloat64Desc(preds, highestLocked)
			continue
		}
		value := f()
		nn := newFloat64NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockFloat64Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Float64MapDesc) Delete(key float64) bool {
	var (
		nodeToDelete *float64NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*float64NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockFloat64Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockFloat64Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Float64MapDesc) Range(f func(key float64, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Float64MapDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Int32Map represents a map based on skip list in ascending order.
type Int32Map struct {
	header       *int32Node
	length       int64
	highestLevel int64 // highest level for now
}

type int32Node struct {
	key   int32
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*int32Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newInt32Node(key int32, value interface{}, level int) *int32Node {
	node := &int32Node{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *int32Node) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *int32Node) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *int32Node) lessthan(key int32) bool {
	return n.key < key
}

func (n *int32Node) equal(key int32) bool {
	return n.key == key
}

// NewInt32 return an empty int32 skipmap.
func NewInt32() *Int32Map {
	h := newInt32Node(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Int32Map{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Int32Map) findNode(key int32, preds *[maxLevel]*int32Node, succs *[maxLevel]*int32Node) *int32Node {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Int32Map) findNodeDelete(key int32, preds *[maxLevel]*int32Node, succs *[maxLevel]*int32Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Int32Map) Store(key int32, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*int32Node
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newInt32Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Int32Map) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Int32Map) Load(key int32) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Int32Map) LoadAndDelete(key int32) (value interface{}, loaded bool) {
	var (
		nodeToDelete *int32Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int32Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockInt32(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Int32Map) LoadOrStore(key int32, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*int32Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int32Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newInt32Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Int32Map) LoadOrStoreLazy(key int32, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*int32Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int32Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockInt32(preds, highestLocked)
			continue
		}
		value := f()
		nn := newInt32Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Int32Map) Delete(key int32) bool {
	var (
		nodeToDelete *int32Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int32Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockInt32(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockInt32(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Int32Map) Range(f func(key int32, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Int32Map) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Int32MapDesc represents a map based on skip list in descending order.
type Int32MapDesc struct {
	header       *int32NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type int32NodeDesc struct {
	key   int32
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*int32NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newInt32NodeDesc(key int32, value interface{}, level int) *int32NodeDesc {
	node := &int32NodeDesc{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *int32NodeDesc) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *int32NodeDesc) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *int32NodeDesc) lessthan(key int32) bool {
	return n.key > key
}

func (n *int32NodeDesc) equal(key int32) bool {
	return n.key == key
}

// NewInt32Desc return an empty int32 skipmap.
func NewInt32Desc() *Int32MapDesc {
	h := newInt32NodeDesc(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Int32MapDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Int32MapDesc) findNode(key int32, preds *[maxLevel]*int32NodeDesc, succs *[maxLevel]*int32NodeDesc) *int32NodeDesc {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Int32MapDesc) findNodeDelete(key int32, preds *[maxLevel]*int32NodeDesc, succs *[maxLevel]*int32NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Int32MapDesc) Store(key int32, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*int32NodeDesc
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newInt32NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Int32MapDesc) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Int32MapDesc) Load(key int32) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Int32MapDesc) LoadAndDelete(key int32) (value interface{}, loaded bool) {
	var (
		nodeToDelete *int32NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int32NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockInt32Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Int32MapDesc) LoadOrStore(key int32, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*int32NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int32NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newInt32NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Int32MapDesc) LoadOrStoreLazy(key int32, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*int32NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int32NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockInt32Desc(preds, highestLocked)
			continue
		}
		value := f()
		nn := newInt32NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Int32MapDesc) Delete(key int32) bool {
	var (
		nodeToDelete *int32NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int32NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockInt32Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockInt32Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Int32MapDesc) Range(f func(key int32, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Int32MapDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Int16Map represents a map based on skip list in ascending order.
type Int16Map struct {
	header       *int16Node
	length       int64
	highestLevel int64 // highest level for now
}

type int16Node struct {
	key   int16
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*int16Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newInt16Node(key int16, value interface{}, level int) *int16Node {
	node := &int16Node{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *int16Node) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *int16Node) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *int16Node) lessthan(key int16) bool {
	return n.key < key
}

func (n *int16Node) equal(key int16) bool {
	return n.key == key
}

// NewInt16 return an empty int16 skipmap.
func NewInt16() *Int16Map {
	h := newInt16Node(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Int16Map{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Int16Map) findNode(key int16, preds *[maxLevel]*int16Node, succs *[maxLevel]*int16Node) *int16Node {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Int16Map) findNodeDelete(key int16, preds *[maxLevel]*int16Node, succs *[maxLevel]*int16Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Int16Map) Store(key int16, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*int16Node
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newInt16Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt16(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Int16Map) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Int16Map) Load(key int16) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Int16Map) LoadAndDelete(key int16) (value interface{}, loaded bool) {
	var (
		nodeToDelete *int16Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int16Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockInt16(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Int16Map) LoadOrStore(key int16, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*int16Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int16Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newInt16Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt16(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Int16Map) LoadOrStoreLazy(key int16, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*int16Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int16Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockInt16(preds, highestLocked)
			continue
		}
		value := f()
		nn := newInt16Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt16(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Int16Map) Delete(key int16) bool {
	var (
		nodeToDelete *int16Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int16Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockInt16(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockInt16(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Int16Map) Range(f func(key int16, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Int16Map) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Int16MapDesc represents a map based on skip list in descending order.
type Int16MapDesc struct {
	header       *int16NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type int16NodeDesc struct {
	key   int16
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*int16NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newInt16NodeDesc(key int16, value interface{}, level int) *int16NodeDesc {
	node := &int16NodeDesc{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *int16NodeDesc) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *int16NodeDesc) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *int16NodeDesc) lessthan(key int16) bool {
	return n.key > key
}

func (n *int16NodeDesc) equal(key int16) bool {
	return n.key == key
}

// NewInt16Desc return an empty int16 skipmap.
func NewInt16Desc() *Int16MapDesc {
	h := newInt16NodeDesc(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Int16MapDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Int16MapDesc) findNode(key int16, preds *[maxLevel]*int16NodeDesc, succs *[maxLevel]*int16NodeDesc) *int16NodeDesc {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Int16MapDesc) findNodeDelete(key int16, preds *[maxLevel]*int16NodeDesc, succs *[maxLevel]*int16NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Int16MapDesc) Store(key int16, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*int16NodeDesc
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newInt16NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt16Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Int16MapDesc) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Int16MapDesc) Load(key int16) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Int16MapDesc) LoadAndDelete(key int16) (value interface{}, loaded bool) {
	var (
		nodeToDelete *int16NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int16NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockInt16Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Int16MapDesc) LoadOrStore(key int16, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*int16NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int16NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newInt16NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt16Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Int16MapDesc) LoadOrStoreLazy(key int16, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*int16NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *int16NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockInt16Desc(preds, highestLocked)
			continue
		}
		value := f()
		nn := newInt16NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt16Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Int16MapDesc) Delete(key int16) bool {
	var (
		nodeToDelete *int16NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*int16NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockInt16Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockInt16Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Int16MapDesc) Range(f func(key int16, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Int16MapDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// IntMap represents a map based on skip list in ascending order.
type IntMap struct {
	header       *intNode
	length       int64
	highestLevel int64 // highest level for now
}

type intNode struct {
	key   int
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*intNode
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newIntNode(key int, value interface{}, level int) *intNode {
	node := &intNode{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *intNode) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *intNode) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *intNode) lessthan(key int) bool {
	return n.key < key
}

func (n *intNode) equal(key int) bool {
	return n.key == key
}

// NewInt return an empty int skipmap.
func NewInt() *IntMap {
	h := newIntNode(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &IntMap{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *IntMap) findNode(key int, preds *[maxLevel]*intNode, succs *[maxLevel]*intNode) *intNode {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *IntMap) findNodeDelete(key int, preds *[maxLevel]*intNode, succs *[maxLevel]*intNode) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *IntMap) Store(key int, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*intNode
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newIntNode(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *IntMap) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *IntMap) Load(key int) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *IntMap) LoadAndDelete(key int) (value interface{}, loaded bool) {
	var (
		nodeToDelete *intNode
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*intNode
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockInt(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *IntMap) LoadOrStore(key int, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*intNode
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *intNode
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newIntNode(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *IntMap) LoadOrStoreLazy(key int, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*intNode
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *intNode
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockInt(preds, highestLocked)
			continue
		}
		value := f()
		nn := newIntNode(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockInt(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *IntMap) Delete(key int) bool {
	var (
		nodeToDelete *intNode
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*intNode
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockInt(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockInt(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *IntMap) Range(f func(key int, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *IntMap) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// IntMapDesc represents a map based on skip list in descending order.
type IntMapDesc struct {
	header       *intNodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type intNodeDesc struct {
	key   int
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*intNodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newIntNodeDesc(key int, value interface{}, level int) *intNodeDesc {
	node := &intNodeDesc{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *intNodeDesc) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *intNodeDesc) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *intNodeDesc) lessthan(key int) bool {
	return n.key > key
}

func (n *intNodeDesc) equal(key int) bool {
	return n.key == key
}

// NewIntDesc return an empty int skipmap.
func NewIntDesc() *IntMapDesc {
	h := newIntNodeDesc(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &IntMapDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *IntMapDesc) findNode(key int, preds *[maxLevel]*intNodeDesc, succs *[maxLevel]*intNodeDesc) *intNodeDesc {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *IntMapDesc) findNodeDelete(key int, preds *[maxLevel]*intNodeDesc, succs *[maxLevel]*intNodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *IntMapDesc) Store(key int, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*intNodeDesc
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newIntNodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockIntDesc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *IntMapDesc) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *IntMapDesc) Load(key int) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *IntMapDesc) LoadAndDelete(key int) (value interface{}, loaded bool) {
	var (
		nodeToDelete *intNodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*intNodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockIntDesc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *IntMapDesc) LoadOrStore(key int, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*intNodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *intNodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newIntNodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockIntDesc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *IntMapDesc) LoadOrStoreLazy(key int, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*intNodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *intNodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockIntDesc(preds, highestLocked)
			continue
		}
		value := f()
		nn := newIntNodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockIntDesc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *IntMapDesc) Delete(key int) bool {
	var (
		nodeToDelete *intNodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*intNodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockIntDesc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockIntDesc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *IntMapDesc) Range(f func(key int, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *IntMapDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint64Map represents a map based on skip list in ascending order.
type Uint64Map struct {
	header       *uint64Node
	length       int64
	highestLevel int64 // highest level for now
}

type uint64Node struct {
	key   uint64
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*uint64Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUuint64Node(key uint64, value interface{}, level int) *uint64Node {
	node := &uint64Node{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint64Node) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *uint64Node) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *uint64Node) lessthan(key uint64) bool {
	return n.key < key
}

func (n *uint64Node) equal(key uint64) bool {
	return n.key == key
}

// NewUint64 return an empty uint64 skipmap.
func NewUint64() *Uint64Map {
	h := newUuint64Node(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint64Map{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Uint64Map) findNode(key uint64, preds *[maxLevel]*uint64Node, succs *[maxLevel]*uint64Node) *uint64Node {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Uint64Map) findNodeDelete(key uint64, preds *[maxLevel]*uint64Node, succs *[maxLevel]*uint64Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Uint64Map) Store(key uint64, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint64Node
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newUuint64Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint64(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Uint64Map) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Uint64Map) Load(key uint64) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Uint64Map) LoadAndDelete(key uint64) (value interface{}, loaded bool) {
	var (
		nodeToDelete *uint64Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint64Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint64(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Uint64Map) LoadOrStore(key uint64, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint64Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint64Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newUuint64Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint64(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Uint64Map) LoadOrStoreLazy(key uint64, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint64Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint64Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockUint64(preds, highestLocked)
			continue
		}
		value := f()
		nn := newUuint64Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint64(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Uint64Map) Delete(key uint64) bool {
	var (
		nodeToDelete *uint64Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint64Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockUint64(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint64(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Uint64Map) Range(f func(key uint64, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Uint64Map) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint64MapDesc represents a map based on skip list in descending order.
type Uint64MapDesc struct {
	header       *uint64NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type uint64NodeDesc struct {
	key   uint64
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*uint64NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUuint64NodeDescDesc(key uint64, value interface{}, level int) *uint64NodeDesc {
	node := &uint64NodeDesc{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint64NodeDesc) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *uint64NodeDesc) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *uint64NodeDesc) lessthan(key uint64) bool {
	return n.key > key
}

func (n *uint64NodeDesc) equal(key uint64) bool {
	return n.key == key
}

// NewUint64Desc return an empty uint64 skipmap.
func NewUint64Desc() *Uint64MapDesc {
	h := newUuint64NodeDescDesc(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint64MapDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Uint64MapDesc) findNode(key uint64, preds *[maxLevel]*uint64NodeDesc, succs *[maxLevel]*uint64NodeDesc) *uint64NodeDesc {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Uint64MapDesc) findNodeDelete(key uint64, preds *[maxLevel]*uint64NodeDesc, succs *[maxLevel]*uint64NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Uint64MapDesc) Store(key uint64, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint64NodeDesc
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newUuint64NodeDescDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint64Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Uint64MapDesc) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Uint64MapDesc) Load(key uint64) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Uint64MapDesc) LoadAndDelete(key uint64) (value interface{}, loaded bool) {
	var (
		nodeToDelete *uint64NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint64NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint64Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Uint64MapDesc) LoadOrStore(key uint64, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint64NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint64NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newUuint64NodeDescDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint64Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Uint64MapDesc) LoadOrStoreLazy(key uint64, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint64NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint64NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockUint64Desc(preds, highestLocked)
			continue
		}
		value := f()
		nn := newUuint64NodeDescDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint64Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Uint64MapDesc) Delete(key uint64) bool {
	var (
		nodeToDelete *uint64NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint64NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockUint64Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint64Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Uint64MapDesc) Range(f func(key uint64, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Uint64MapDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint32Map represents a map based on skip list in ascending order.
type Uint32Map struct {
	header       *uint32Node
	length       int64
	highestLevel int64 // highest level for now
}

type uint32Node struct {
	key   uint32
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*uint32Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUint32Node(key uint32, value interface{}, level int) *uint32Node {
	node := &uint32Node{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint32Node) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *uint32Node) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *uint32Node) lessthan(key uint32) bool {
	return n.key < key
}

func (n *uint32Node) equal(key uint32) bool {
	return n.key == key
}

// NewUint32 return an empty uint32 skipmap.
func NewUint32() *Uint32Map {
	h := newUint32Node(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint32Map{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Uint32Map) findNode(key uint32, preds *[maxLevel]*uint32Node, succs *[maxLevel]*uint32Node) *uint32Node {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Uint32Map) findNodeDelete(key uint32, preds *[maxLevel]*uint32Node, succs *[maxLevel]*uint32Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Uint32Map) Store(key uint32, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint32Node
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newUint32Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Uint32Map) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Uint32Map) Load(key uint32) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Uint32Map) LoadAndDelete(key uint32) (value interface{}, loaded bool) {
	var (
		nodeToDelete *uint32Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint32Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint32(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Uint32Map) LoadOrStore(key uint32, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint32Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint32Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newUint32Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Uint32Map) LoadOrStoreLazy(key uint32, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint32Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint32Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockUint32(preds, highestLocked)
			continue
		}
		value := f()
		nn := newUint32Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint32(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Uint32Map) Delete(key uint32) bool {
	var (
		nodeToDelete *uint32Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint32Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockUint32(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint32(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Uint32Map) Range(f func(key uint32, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Uint32Map) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint32MapDesc represents a map based on skip list in descending order.
type Uint32MapDesc struct {
	header       *uint32NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type uint32NodeDesc struct {
	key   uint32
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*uint32NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUint32NodeDesc(key uint32, value interface{}, level int) *uint32NodeDesc {
	node := &uint32NodeDesc{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint32NodeDesc) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *uint32NodeDesc) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *uint32NodeDesc) lessthan(key uint32) bool {
	return n.key > key
}

func (n *uint32NodeDesc) equal(key uint32) bool {
	return n.key == key
}

// NewUint32Desc return an empty uint32 skipmap.
func NewUint32Desc() *Uint32MapDesc {
	h := newUint32NodeDesc(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint32MapDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Uint32MapDesc) findNode(key uint32, preds *[maxLevel]*uint32NodeDesc, succs *[maxLevel]*uint32NodeDesc) *uint32NodeDesc {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Uint32MapDesc) findNodeDelete(key uint32, preds *[maxLevel]*uint32NodeDesc, succs *[maxLevel]*uint32NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Uint32MapDesc) Store(key uint32, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint32NodeDesc
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newUint32NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Uint32MapDesc) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Uint32MapDesc) Load(key uint32) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Uint32MapDesc) LoadAndDelete(key uint32) (value interface{}, loaded bool) {
	var (
		nodeToDelete *uint32NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint32NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint32Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Uint32MapDesc) LoadOrStore(key uint32, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint32NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint32NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newUint32NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Uint32MapDesc) LoadOrStoreLazy(key uint32, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint32NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint32NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockUint32Desc(preds, highestLocked)
			continue
		}
		value := f()
		nn := newUint32NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint32Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Uint32MapDesc) Delete(key uint32) bool {
	var (
		nodeToDelete *uint32NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint32NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockUint32Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint32Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Uint32MapDesc) Range(f func(key uint32, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Uint32MapDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint16Map represents a map based on skip list in ascending order.
type Uint16Map struct {
	header       *uint16Node
	length       int64
	highestLevel int64 // highest level for now
}

type uint16Node struct {
	key   uint16
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*uint16Node
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUint16Node(key uint16, value interface{}, level int) *uint16Node {
	node := &uint16Node{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint16Node) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *uint16Node) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *uint16Node) lessthan(key uint16) bool {
	return n.key < key
}

func (n *uint16Node) equal(key uint16) bool {
	return n.key == key
}

// NewUint16 return an empty uint16 skipmap.
func NewUint16() *Uint16Map {
	h := newUint16Node(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint16Map{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Uint16Map) findNode(key uint16, preds *[maxLevel]*uint16Node, succs *[maxLevel]*uint16Node) *uint16Node {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Uint16Map) findNodeDelete(key uint16, preds *[maxLevel]*uint16Node, succs *[maxLevel]*uint16Node) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Uint16Map) Store(key uint16, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint16Node
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newUint16Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint16(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Uint16Map) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Uint16Map) Load(key uint16) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Uint16Map) LoadAndDelete(key uint16) (value interface{}, loaded bool) {
	var (
		nodeToDelete *uint16Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint16Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint16(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Uint16Map) LoadOrStore(key uint16, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint16Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint16Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newUint16Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint16(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Uint16Map) LoadOrStoreLazy(key uint16, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint16Node
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint16Node
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockUint16(preds, highestLocked)
			continue
		}
		value := f()
		nn := newUint16Node(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint16(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Uint16Map) Delete(key uint16) bool {
	var (
		nodeToDelete *uint16Node
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint16Node
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockUint16(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint16(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Uint16Map) Range(f func(key uint16, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Uint16Map) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Uint16MapDesc represents a map based on skip list in descending order.
type Uint16MapDesc struct {
	header       *uint16NodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type uint16NodeDesc struct {
	key   uint16
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*uint16NodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUint16NodeDesc(key uint16, value interface{}, level int) *uint16NodeDesc {
	node := &uint16NodeDesc{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uint16NodeDesc) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *uint16NodeDesc) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *uint16NodeDesc) lessthan(key uint16) bool {
	return n.key > key
}

func (n *uint16NodeDesc) equal(key uint16) bool {
	return n.key == key
}

// NewUint16Desc return an empty uint16 skipmap.
func NewUint16Desc() *Uint16MapDesc {
	h := newUint16NodeDesc(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &Uint16MapDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *Uint16MapDesc) findNode(key uint16, preds *[maxLevel]*uint16NodeDesc, succs *[maxLevel]*uint16NodeDesc) *uint16NodeDesc {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *Uint16MapDesc) findNodeDelete(key uint16, preds *[maxLevel]*uint16NodeDesc, succs *[maxLevel]*uint16NodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *Uint16MapDesc) Store(key uint16, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uint16NodeDesc
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newUint16NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint16Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *Uint16MapDesc) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *Uint16MapDesc) Load(key uint16) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *Uint16MapDesc) LoadAndDelete(key uint16) (value interface{}, loaded bool) {
	var (
		nodeToDelete *uint16NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint16NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint16Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *Uint16MapDesc) LoadOrStore(key uint16, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint16NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint16NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newUint16NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint16Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *Uint16MapDesc) LoadOrStoreLazy(key uint16, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uint16NodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uint16NodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockUint16Desc(preds, highestLocked)
			continue
		}
		value := f()
		nn := newUint16NodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint16Desc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *Uint16MapDesc) Delete(key uint16) bool {
	var (
		nodeToDelete *uint16NodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uint16NodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockUint16Desc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint16Desc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *Uint16MapDesc) Range(f func(key uint16, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *Uint16MapDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// UintMap represents a map based on skip list in ascending order.
type UintMap struct {
	header       *uintNode
	length       int64
	highestLevel int64 // highest level for now
}

type uintNode struct {
	key   uint
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*uintNode
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUintNode(key uint, value interface{}, level int) *uintNode {
	node := &uintNode{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uintNode) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *uintNode) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *uintNode) lessthan(key uint) bool {
	return n.key < key
}

func (n *uintNode) equal(key uint) bool {
	return n.key == key
}

// NewUint return an empty uint skipmap.
func NewUint() *UintMap {
	h := newUintNode(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &UintMap{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *UintMap) findNode(key uint, preds *[maxLevel]*uintNode, succs *[maxLevel]*uintNode) *uintNode {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *UintMap) findNodeDelete(key uint, preds *[maxLevel]*uintNode, succs *[maxLevel]*uintNode) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *UintMap) Store(key uint, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uintNode
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newUintNode(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *UintMap) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *UintMap) Load(key uint) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *UintMap) LoadAndDelete(key uint) (value interface{}, loaded bool) {
	var (
		nodeToDelete *uintNode
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uintNode
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *UintMap) LoadOrStore(key uint, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uintNode
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uintNode
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newUintNode(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *UintMap) LoadOrStoreLazy(key uint, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uintNode
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uintNode
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockUint(preds, highestLocked)
			continue
		}
		value := f()
		nn := newUintNode(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUint(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *UintMap) Delete(key uint) bool {
	var (
		nodeToDelete *uintNode
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uintNode
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockUint(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUint(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *UintMap) Range(f func(key uint, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *UintMap) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// UintMapDesc represents a map based on skip list in descending order.
type UintMapDesc struct {
	header       *uintNodeDesc
	length       int64
	highestLevel int64 // highest level for now
}

type uintNodeDesc struct {
	key   uint
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*uintNodeDesc
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newUintNodeDesc(key uint, value interface{}, level int) *uintNodeDesc {
	node := &uintNodeDesc{
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *uintNodeDesc) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *uintNodeDesc) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

func (n *uintNodeDesc) lessthan(key uint) bool {
	return n.key > key
}

func (n *uintNodeDesc) equal(key uint) bool {
	return n.key == key
}

// NewUintDesc return an empty uint skipmap.
func NewUintDesc() *UintMapDesc {
	h := newUintNodeDesc(0, "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &UintMapDesc{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *UintMapDesc) findNode(key uint, preds *[maxLevel]*uintNodeDesc, succs *[maxLevel]*uintNodeDesc) *uintNodeDesc {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.equal(key) {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *UintMapDesc) findNodeDelete(key uint, preds *[maxLevel]*uintNodeDesc, succs *[maxLevel]*uintNodeDesc) int {
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.lessthan(key) {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.equal(key) {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *UintMapDesc) Store(key uint, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*uintNodeDesc
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newUintNodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUintDesc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *UintMapDesc) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *UintMapDesc) Load(key uint) (value interface{}, ok bool) {
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.lessthan(key) {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.equal(key) {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *UintMapDesc) LoadAndDelete(key uint) (value interface{}, loaded bool) {
	var (
		nodeToDelete *uintNodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uintNodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUintDesc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *UintMapDesc) LoadOrStore(key uint, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uintNodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uintNodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newUintNodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUintDesc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *UintMapDesc) LoadOrStoreLazy(key uint, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*uintNodeDesc
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *uintNodeDesc
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockUintDesc(preds, highestLocked)
			continue
		}
		value := f()
		nn := newUintNodeDesc(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockUintDesc(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *UintMapDesc) Delete(key uint) bool {
	var (
		nodeToDelete *uintNodeDesc
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*uintNodeDesc
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockUintDesc(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockUintDesc(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *UintMapDesc) Range(f func(key uint, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *UintMapDesc) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// StringMap represents a map based on skip list.
type StringMap struct {
	header       *stringNode
	length       int64
	highestLevel int64 // highest level for now
}

type stringNode struct {
	key   string
	score uint64
	value unsafe.Pointer // *interface{}
	next  optionalArray  // [level]*stringNode
	mu    sync.Mutex
	flags bitflag
	level uint32
}

func newStringNode(key string, value interface{}, level int) *stringNode {
	node := &stringNode{
		score: hash(key),
		key:   key,
		level: uint32(level),
	}
	node.storeVal(value)
	if level > op1 {
		node.next.extra = new([op2]unsafe.Pointer)
	}
	return node
}

func (n *stringNode) storeVal(value interface{}) {
	atomic.StorePointer(&n.value, unsafe.Pointer(&value))
}

func (n *stringNode) loadVal() interface{} {
	return *(*interface{})(atomic.LoadPointer(&n.value))
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

// NewString return an empty int64 skipmap.
func NewString() *StringMap {
	h := newStringNode("", "", maxLevel)
	h.flags.SetTrue(fullyLinked)
	return &StringMap{
		header:       h,
		highestLevel: defaultHighestLevel,
	}
}

// findNode takes a key and two maximal-height arrays then searches exactly as in a sequential skipmap.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
// (without fullpath, if find the node will return immediately)
func (s *StringMap) findNode(key string, preds *[maxLevel]*stringNode, succs *[maxLevel]*stringNode) *stringNode {
	score := hash(key)
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.cmp(score, key) < 0 {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skipmap.
		if succ != nil && succ.cmp(score, key) == 0 {
			return succ
		}
	}
	return nil
}

// findNodeDelete takes a key and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > key >= succs[i].
func (s *StringMap) findNodeDelete(key string, preds *[maxLevel]*stringNode, succs *[maxLevel]*stringNode) int {
	score := hash(key)
	// lFound represents the index of the first layer at which it found a node.
	lFound, x := -1, s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		succ := x.atomicLoadNext(i)
		for succ != nil && succ.cmp(score, key) < 0 {
			x = succ
			succ = x.atomicLoadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the key already in the skip list.
		if lFound == -1 && succ != nil && succ.cmp(score, key) == 0 {
			lFound = i
		}
	}
	return lFound
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

// Store sets the value for a key.
func (s *StringMap) Store(key string, value interface{}) {
	level := s.randomlevel()
	var preds, succs [maxLevel]*stringNode
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just replace the value.
				nodeFound.storeVal(value)
				return
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
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

		nn := newStringNode(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockString(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return
	}
}

func (s *StringMap) randomlevel() int {
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

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (s *StringMap) Load(key string) (value interface{}, ok bool) {
	score := hash(key)
	x := s.header
	for i := int(atomic.LoadInt64(&s.highestLevel)) - 1; i >= 0; i-- {
		nex := x.atomicLoadNext(i)
		for nex != nil && nex.cmp(score, key) < 0 {
			x = nex
			nex = x.atomicLoadNext(i)
		}

		// Check if the key already in the skip list.
		if nex != nil && nex.cmp(score, key) == 0 {
			if nex.flags.MGet(fullyLinked|marked, fullyLinked) {
				return nex.loadVal(), true
			}
			return nil, false
		}
	}
	return nil, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// (Modified from Delete)
func (s *StringMap) LoadAndDelete(key string) (value interface{}, loaded bool) {
	var (
		nodeToDelete *stringNode
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*stringNode
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return nil, false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
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
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockString(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return nodeToDelete.loadVal(), true
		}
		return nil, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from Store)
func (s *StringMap) LoadOrStore(key string, value interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*stringNode
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *stringNode
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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

		nn := newStringNode(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockString(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// LoadOrStoreLazy returns the existing value for the key if present.
// Otherwise, it stores and returns the given value from f, f will only be called once.
// The loaded result is true if the value was loaded, false if stored.
// (Modified from LoadOrStore)
func (s *StringMap) LoadOrStoreLazy(key string, f func() interface{}) (actual interface{}, loaded bool) {
	var (
		level        int
		preds, succs [maxLevel]*stringNode
		hl           = int(atomic.LoadInt64(&s.highestLevel))
	)
	for {
		nodeFound := s.findNode(key, &preds, &succs)
		if nodeFound != nil { // indicating the key is already in the skip-list
			if !nodeFound.flags.Get(marked) {
				// We don't need to care about whether or not the node is fully linked,
				// just return the value.
				return nodeFound.loadVal(), true
			}
			// If the node is marked, represents some other goroutines is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *stringNode
		)
		if level == 0 {
			level = s.randomlevel()
			if level > hl {
				// If the highest level is updated, usually means that many goroutines
				// are inserting items. Hopefully we can find a better path in next loop.
				// TODO(zyh): consider filling the preds if s.header[level].next == nil,
				// but this strategy's performance is almost the same as the existing method.
				continue
			}
		}
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
			valid = !pred.flags.Get(marked) && pred.loadNext(layer) == succ && (succ == nil || !succ.flags.Get(marked))
		}
		if !valid {
			unlockString(preds, highestLocked)
			continue
		}
		value := f()
		nn := newStringNode(key, value, level)
		for layer := 0; layer < level; layer++ {
			nn.storeNext(layer, succs[layer])
			preds[layer].atomicStoreNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockString(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return value, false
	}
}

// Delete deletes the value for a key.
func (s *StringMap) Delete(key string) bool {
	var (
		nodeToDelete *stringNode
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*stringNode
	)
	for {
		lFound := s.findNodeDelete(key, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (int(succs[lFound].level)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToDelete = succs[lFound]
				topLayer = lFound
				nodeToDelete.mu.Lock()
				if nodeToDelete.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToDelete.mu.Unlock()
					return false
				}
				nodeToDelete.flags.SetTrue(marked)
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
				// during this process, or the previous is deleted by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.atomicLoadNext(layer) == succ
			}
			if !valid {
				unlockString(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToDelete`, no other goroutine will modify it.
				// So we don't need `nodeToDelete.loadNext`
				preds[i].atomicStoreNext(i, nodeToDelete.loadNext(i))
			}
			nodeToDelete.mu.Unlock()
			unlockString(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each key and value present in the skipmap.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
func (s *StringMap) Range(f func(key string, value interface{}) bool) {
	x := s.header.atomicLoadNext(0)
	for x != nil {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.atomicLoadNext(0)
			continue
		}
		if !f(x.key, x.loadVal()) {
			break
		}
		x = x.atomicLoadNext(0)
	}
}

// Len return the length of this skipmap.
func (s *StringMap) Len() int {
	return int(atomic.LoadInt64(&s.length))
}

// Return 1 if n is bigger, 0 if equal, else -1.
func (n *stringNode) cmp(score uint64, key string) int {
	if n.score > score {
		return 1
	} else if n.score == score {
		return cmpstring(n.key, key)
	}
	return -1
}
