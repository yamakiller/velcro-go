package loadbalance

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/rpc/pkg/discovery"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/hashx3"
)

// KeyFunc 应该返回一个非空字符串,代表给定上下文中的请求.
type KeyFunc func(ctx context.Context, request interface{}) string

// ConsistentHashOption.
type ConsistentHashOption struct {
	GetKey KeyFunc

	// 如果设置, 当连接到主​​节点失败时将使用副本;
	// 这会带来额外的内存和CPU成本;
	// 如果不设置, 连接失败时会立即返回错误.
	Replica uint32

	// 每个真实节点对应的虚拟节点数量;
	// 该值越大, 内存和计算成本越高,负载越均衡; 当节点数量较多时, 可以设置
	// 小一些; 反之可以设大中值 VirtualFactor * Weight(如果 Weighted 为
	// true) 建议在 1000 左右 建议虚拟节点总数在 2000W 以内( 1000W情况下
	// 构建一次需要 250ms,但理论上是可以在 3 秒内在后台构建).
	VirtualFactor uint32

	// 是否按照Weight进行负载均衡
	// 如果为 false, 则忽略每个实例的 Weight, 并生成 VirtualFactor 虚拟节点以进行无差别负载平衡.
	// 如果为 true,  则为每个实例生成 Weight() * VirtualFactor 虚拟节点.
	// 注意，例如权重为0时, 无论VirtualFactor数量是多少，都不会生成虚拟节点. 建议将其设置为true,
	// 但要注意适当减少VirtualFactor.
	Weighted bool

	// 是否进行过期处理
	// 该实现将缓存所有密钥
	// 如果永不过期, 可能会导致内存不断增长, 最终 OOM 设置过期将导致额外的性能开销
	// 当前实现每分钟扫描一次删除, 当实例更改重建时删除一次建议始终设置该值不小于两分钟
	ExpireDuration time.Duration
}

// NewConsistentHashOption 创建一个默认的 ConsistentHashOption.
func NewConsistentHashOption(f KeyFunc) ConsistentHashOption {
	return ConsistentHashOption{
		GetKey:         f,
		Replica:        0,
		VirtualFactor:  100,
		Weighted:       true,
		ExpireDuration: 2 * time.Minute,
	}
}

var (
	consistPickerPool         sync.Pool
	consistBalancers          []*consistBalancer
	consistBalancersLock      sync.RWMutex
	consistBalancerDaemonOnce sync.Once
)

func init() {
	consistPickerPool.New = newConsistPicker
}

type virtualNode struct {
	hash     uint64
	RealNode *realNode
}

type realNode struct {
	Ins discovery.Instance
}

type consistResult struct {
	Primary  discovery.Instance
	Replicas []discovery.Instance
	Touch    atomic.Value
}

type consistInfo struct {
	cachedConsistResult sync.Map
	sfg                 singleflight.Group

	realNodes    []realNode
	virtualNodes []virtualNode
}

type vNodeType struct {
	s []virtualNode
}

func (v *vNodeType) Len() int {
	return len(v.s)
}

func (v *vNodeType) Less(i, j int) bool {
	return v.s[i].hash < v.s[j].hash
}

func (v *vNodeType) Swap(i, j int) {
	v.s[i], v.s[j] = v.s[j], v.s[i]
}

type consistPicker struct {
	cb     *consistBalancer
	info   *consistInfo
	index  int
	result *consistResult
}

func newConsistPicker() interface{} {
	return &consistPicker{}
}

func (cp *consistPicker) zero() {
	cp.info = nil
	cp.cb = nil
	cp.index = 0
	cp.result = nil
}

func (cp *consistPicker) Recycle() {
	cp.zero()
	consistPickerPool.Put(cp)
}

// Next is not concurrency safe.
func (cp *consistPicker) Next(ctx context.Context, request interface{}) discovery.Instance {
	if len(cp.info.realNodes) == 0 {
		return nil
	}

	if cp.result == nil {
		key := cp.cb.opt.GetKey(ctx, request)
		if key == "" {
			return nil
		}
		cp.result = cp.getConsistResult(hashx3.HashString(key))
		cp.index = 0
		return cp.result.Primary
	}
	if cp.index < len(cp.result.Replicas) {
		cp.index++
		return cp.result.Replicas[cp.index-1]
	}
	return nil
}

func (cp *consistPicker) getConsistResult(key uint64) *consistResult {
	var cr *consistResult
	cri, ok := cp.info.cachedConsistResult.Load(key)
	if !ok {
		cri, _, _ = cp.info.sfg.Do(strconv.FormatUint(key, 10), func() (interface{}, error) {
			cr := buildConsistResult(cp.cb, cp.info, key)
			if cp.cb.opt.ExpireDuration > 0 {
				cr.Touch.Store(time.Now())
			}
			return cr, nil
		})
		cp.info.cachedConsistResult.Store(key, cri)
	}
	cr = cri.(*consistResult)
	if cp.cb.opt.ExpireDuration > 0 {
		cr.Touch.Store(time.Now())
	}
	return cr
}

func buildConsistResult(cb *consistBalancer, info *consistInfo, key uint64) *consistResult {
	cr := &consistResult{}
	index := sort.Search(len(info.virtualNodes), func(i int) bool {
		return info.virtualNodes[i].hash > key
	})

	if index == len(info.virtualNodes) {
		index = 0
	}
	cr.Primary = info.virtualNodes[index].RealNode.Ins
	replicas := int(cb.opt.Replica)
	// remove the primary node
	if len(info.realNodes)-1 < replicas {
		replicas = len(info.realNodes) - 1
	}

	if replicas > 0 {
		used := make(map[discovery.Instance]struct{}, replicas) // should be 1 + replicas - 1
		used[cr.Primary] = struct{}{}
		cr.Replicas = make([]discovery.Instance, replicas)
		for i := 0; i < replicas; i++ {
			// 查找下一个未使用的实例
			// 副本之前已调整过，因此我们可以保证找到一个
			for {
				index++
				if index == len(info.virtualNodes) {
					index = 0
				}
				ins := info.virtualNodes[index].RealNode.Ins
				if _, ok := used[ins]; !ok {
					used[ins] = struct{}{}
					cr.Replicas[i] = ins
					break
				}
			}
		}
	}
	return cr
}

type consistBalancer struct {
	cachedConsistInfo sync.Map
	// 这个锁的主要目的是为了提高性能，防止Change在过期时执行
	// 这可能会导致Change做大量额外的计算和内存分配
	updateLock sync.Mutex
	opt        ConsistentHashOption
	sfg        singleflight.Group
}

// NewConsistBalancer 使用给定选项创建一个新的一致性平衡器.
func NewConsistBalancer(opt ConsistentHashOption) Loadbalancer {
	if opt.GetKey == nil {
		panic("loadbalancer: new consistBalancer failed, getKey func cannot be nil")
	}
	if opt.VirtualFactor == 0 {
		panic("loadbalancer: new consistBalancer failed, virtual factor must > 0")
	}
	cb := &consistBalancer{
		opt: opt,
	}
	if cb.opt.ExpireDuration > 0 {
		cb.AddToDaemon()
	}
	return cb
}

// AddToDaemon 向守护进程过期例程添加一个平衡器.
func (cb *consistBalancer) AddToDaemon() {
	// do delete func
	consistBalancerDaemonOnce.Do(func() {
		gofunc.GoFunc(context.Background(), func() {
			for range time.Tick((2 * time.Minute)) {
				consistBalancersLock.RLock()
				now := time.Now()
				for _, lb := range consistBalancers {
					if lb.opt.ExpireDuration > 0 {
						lb.updateLock.Lock()
						lb.cachedConsistInfo.Range(func(key, value interface{}) bool {
							ci := value.(*consistInfo)
							ci.cachedConsistResult.Range(func(key, value interface{}) bool {
								t := value.(*consistResult).Touch.Load().(time.Time)
								if now.After(t.Add(cb.opt.ExpireDuration)) {
									ci.cachedConsistResult.Delete(key)
								}
								return true
							})
							return true
						})
						lb.updateLock.Unlock()
					}
				}
				consistBalancersLock.RUnlock()
			}
		})
	})

	consistBalancersLock.Lock()
	consistBalancers = append(consistBalancers, cb)
	consistBalancersLock.Unlock()
}

// GetPicker 实现Loadbalancer接口.
func (cb *consistBalancer) GetPicker(e discovery.Result) Picker {
	var ci *consistInfo
	if e.Cacheable {
		cii, ok := cb.cachedConsistInfo.Load(e.CacheKey)
		if !ok {
			cii, _, _ = cb.sfg.Do(e.CacheKey, func() (interface{}, error) {
				return cb.newConsistInfo(e), nil
			})
			cb.cachedConsistInfo.Store(e.CacheKey, cii)
		}
		ci = cii.(*consistInfo)
	} else {
		ci = cb.newConsistInfo(e)
	}

	picker := consistPickerPool.Get().(*consistPicker)
	picker.cb = cb
	picker.info = ci

	return picker
}

func (cb *consistBalancer) newConsistInfo(e discovery.Result) *consistInfo {
	ci := &consistInfo{}
	ci.realNodes, ci.virtualNodes = cb.buildNodes(e.Instances)
	return ci
}

func (cb *consistBalancer) buildNodes(ins []discovery.Instance) ([]realNode, []virtualNode) {
	ret := make([]realNode, len(ins))
	for i := range ins {
		ret[i].Ins = ins[i]
	}
	return ret, cb.buildVirtualNodes(ret)
}

func (cb *consistBalancer) buildVirtualNodes(rNodes []realNode) []virtualNode {
	totalLen := 0
	for i := range rNodes {
		totalLen += cb.getVirtualNodeLen(rNodes[i])
	}

	ret := make([]virtualNode, totalLen)
	if totalLen == 0 {
		return ret
	}
	maxLen, maxSerial := 0, 0
	for i := range rNodes {
		if len(rNodes[i].Ins.Address().String()) > maxLen {
			maxLen = len(rNodes[i].Ins.Address().String())
		}
		if vNodeLen := cb.getVirtualNodeLen(rNodes[i]); vNodeLen > maxSerial {
			maxSerial = vNodeLen
		}
	}
	l := maxLen + 1 + utils.GetUIntLen(uint64(maxSerial)) // "$address + # + itoa(i)"
	// pre-allocate []byte here, and reuse it to prevent memory allocation.
	b := make([]byte, l)

	// record the start index.
	cur := 0
	for i := range rNodes {
		bAddr := utils.StringToSliceByte(rNodes[i].Ins.Address().String())
		// Assign the first few bits of b to string.
		copy(b, bAddr)

		// Initialize the last few bits, skipping '#'.
		for j := len(bAddr) + 1; j < len(b); j++ {
			b[j] = 0
		}
		b[len(bAddr)] = '#'

		vLen := cb.getVirtualNodeLen(rNodes[i])
		for j := 0; j < vLen; j++ {
			k := j
			cnt := 0
			// Assign values to b one by one, starting with the last one.
			for k > 0 {
				b[l-1-cnt] = byte(k % 10)
				k /= 10
				cnt++
			}
			// At this point, the index inside ret should be cur + j.
			index := cur + j
			ret[index].hash = hashx3.Hash(b)
			ret[index].RealNode = &rNodes[i]
		}
		cur += vLen
	}
	sort.Sort(&vNodeType{s: ret})
	return ret
}

// 从一个realNode获取虚拟节点号.
// 如果 cb.opt.Weighted 选项为 false, 乘数为 1,虚拟节点数等于 VirtualFactor.
func (cb *consistBalancer) getVirtualNodeLen(rNode realNode) int {
	if cb.opt.Weighted {
		return rNode.Ins.Weight() * int(cb.opt.VirtualFactor)
	}
	return int(cb.opt.VirtualFactor)
}

func (cb *consistBalancer) updateConsistInfo(e discovery.Result) {
	newInfo := cb.newConsistInfo(e)
	infoI, loaded := cb.cachedConsistInfo.LoadOrStore(e.CacheKey, newInfo)
	if !loaded {
		return
	}
	info := infoI.(*consistInfo)
	// Warm up.
	// The reason for not modifying info directly is that there is no guarantee of concurrency security.
	info.cachedConsistResult.Range(func(key, value interface{}) bool {
		cr := buildConsistResult(cb, newInfo, key.(uint64))
		if cb.opt.ExpireDuration > 0 {
			t := value.(*consistResult).Touch.Load().(time.Time)
			if time.Now().After(t.Add(cb.opt.ExpireDuration)) {
				return true
			}
			cr.Touch.Store(t)
		}
		newInfo.cachedConsistResult.Store(key, cr)
		return true
	})
	cb.cachedConsistInfo.Store(e.CacheKey, newInfo)
}

// Rebalance 实现Rebalancer接口.
func (cb *consistBalancer) Rebalance(change discovery.Change) {
	if !change.Result.Cacheable {
		return
	}
	// TODO: 可以使用红黑树进行性优化;
	// 当前使用全构建
	cb.updateLock.Lock()
	cb.updateConsistInfo(change.Result)
	cb.updateLock.Unlock()
}

// Delete 实现Rebalancer接口.
func (cb *consistBalancer) Delete(change discovery.Change) {
	if !change.Result.Cacheable {
		return
	}
	// FIXME: If Delete and Rebalance occur together (Discovery OnDelete and OnChange are triggered at the same time),
	// it may cause the delete to fail and eventually lead to a resource leak.
	cb.updateLock.Lock()
	cb.cachedConsistInfo.Delete(change.Result.CacheKey)
	cb.updateLock.Unlock()
}

func (cb *consistBalancer) Name() string {
	return "consist"
}
