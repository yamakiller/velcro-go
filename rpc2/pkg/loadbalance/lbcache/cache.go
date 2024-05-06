package lbcache

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/rpc/utils/diagnosis"
	"github.com/yamakiller/velcro-go/rpc2/pkg/discovery"
	"github.com/yamakiller/velcro-go/rpc2/pkg/loadbalance"
	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/vlog"
	"golang.org/x/sync/singleflight"
)

const (
	defaultRefreshInterval = 5 * time.Second
	defaultExpireInterval  = 15 * time.Second
)

var (
	balancerFactories    sync.Map // key: resolver name + loadbalance name
	balancerFactoriesSfg singleflight.Group
)

// Options 用于创建构建器
type Options struct {
	// 及时刷新发现结果
	RefreshInterval time.Duration

	// Balancer 过期检查间隔
	// 我们需要删除空闲的 Balancer 以节省资源
	ExpireInterval time.Duration

	// DiagnosisService 用于诊断的寄存器信息
	DiagnosisService diagnosis.Service

	// Cacheable 用于指示工厂是否可以在多个客户端之间共享
	Cacheable bool
}

func (v *Options) check() {
	if v.RefreshInterval <= 0 {
		v.RefreshInterval = defaultRefreshInterval
	}
	if v.ExpireInterval <= 0 {
		v.ExpireInterval = defaultExpireInterval
	}
}

// Hookable 为重新平衡器事件添加钩子
type Hookable interface {
	// register 用于重新平衡事件的负载平衡重新平衡挂钩
	RegisterRebalanceHook(func(ch *discovery.Change)) (index int)
	DeregisterRebalanceHook(index int)
	// register 用于删除事件的负载平衡删除挂钩
	RegisterDeleteHook(func(ch *discovery.Change)) (index int)
	DeregisterDeleteHook(index int)
}

// BalancerFactory 获取或创建具有给定目标的平衡器
// 如果它具有相同的密钥(reslover.Target(target)),我们将缓存并重用Balance
type BalancerFactory struct {
	Hookable
	opts       Options
	cache      sync.Map // key -> LoadBalancer
	resolver   discovery.Resolver
	balancer   loadbalance.Loadbalancer
	rebalancer loadbalance.Rebalancer
	sfg        singleflight.Group
}

func cacheKey(resolver, balancer string, opts Options) string {
	return fmt.Sprintf("%s|%s|{%s %s}", resolver, balancer, opts.RefreshInterval, opts.ExpireInterval)
}

func newBalancerFactory(resolver discovery.Resolver, balancer loadbalance.Loadbalancer, opts Options) *BalancerFactory {
	b := &BalancerFactory{
		opts:     opts,
		resolver: resolver,
		balancer: balancer,
	}
	if rb, ok := balancer.(loadbalance.Rebalancer); ok {
		hrb := newHookRebalancer(rb)
		b.rebalancer = hrb
		b.Hookable = hrb
	} else {
		b.Hookable = noopHookRebalancer{}
	}
	go b.watcher()
	return b
}

// NewBalancerFactory 获取或创建平衡器实例缓存键的平衡器工厂,其中包含解析器名称、平衡器名称和选项.
func NewBalancerFactory(resolver discovery.Resolver, balancer loadbalance.Loadbalancer, opts Options) *BalancerFactory {
	opts.check()
	if !opts.Cacheable {
		return newBalancerFactory(resolver, balancer, opts)
	}
	uniqueKey := cacheKey(resolver.Name(), balancer.Name(), opts)
	val, ok := balancerFactories.Load(uniqueKey)
	if ok {
		return val.(*BalancerFactory)
	}
	val, _, _ = balancerFactoriesSfg.Do(uniqueKey, func() (interface{}, error) {
		b := newBalancerFactory(resolver, balancer, opts)
		balancerFactories.Store(uniqueKey, b)
		return b, nil
	})
	return val.(*BalancerFactory)
}

// watch 过期的平衡器
func (b *BalancerFactory) watcher() {
	for range time.Tick(b.opts.ExpireInterval) {
		b.cache.Range(func(key, value interface{}) bool {
			bl := value.(*Balancer)
			if atomic.CompareAndSwapInt32(&bl.expire, 0, 1) {
				// 1. 设置过期标志
				// 2. 等待下一个 ticker 收集, 也许平衡器又被使用了
				// (避免立即删除最近创建的平衡器)
			} else {
				b.cache.Delete(key)
				bl.close()
			}
			return true
		})
	}
}

// 带有解析器名称前缀的缓存键避免平衡器冲突
func renameResultCacheKey(res *discovery.Result, resolverName string) {
	res.CacheKey = resolverName + ":" + res.CacheKey
}

// 如果不存在则创建一个新的平衡器
func (b *BalancerFactory) Get(ctx context.Context, target rpcinfo.EndpointInfo) (*Balancer, error) {
	desc := b.resolver.Target(ctx, target)
	val, ok := b.cache.Load(desc)
	if ok {
		return val.(*Balancer), nil
	}
	val, err, _ := b.sfg.Do(desc, func() (interface{}, error) {
		res, err := b.resolver.Resolve(ctx, desc)
		if err != nil {
			return nil, err
		}
		renameResultCacheKey(&res, b.resolver.Name())
		bl := &Balancer{
			b:      b,
			target: desc,
		}
		bl.res.Store(res)
		bl.sharedTicker = getSharedTicker(bl, b.opts.RefreshInterval)
		b.cache.Store(desc, bl)
		return bl, nil
	})
	if err != nil {
		return nil, err
	}
	return val.(*Balancer), nil
}

// Balancer 与loadbalance.Loadbalancer相同,但没有resolver.
// 已缓存的结果
type Balancer struct {
	b            *BalancerFactory
	target       string       // 从解析器的 Target 方法返回的描述
	res          atomic.Value // 最新和先前的发现结果
	expire       int32        // 0 = normal, 1 = expire and collect next ticker
	sharedTicker *utils.SharedTicker
}

func (bl *Balancer) Refresh() {
	res, err := bl.b.resolver.Resolve(context.Background(), bl.target)
	if err != nil {
		vlog.Warnf("VELCRO: resolver refresh failed, key=%s error=%s", bl.target, err.Error())
		return
	}
	renameResultCacheKey(&res, bl.b.resolver.Name())
	prev := bl.res.Load().(discovery.Result)
	if bl.b.rebalancer != nil {
		if ch, ok := bl.b.resolver.Diff(res.CacheKey, prev, res); ok {
			bl.b.rebalancer.Rebalance(ch)
		}
	}
	// 替换之前的结果
	bl.res.Store(res)
}

// Tick i实现接口 utils.TickerTask.
func (bl *Balancer) Tick() {
	bl.Refresh()
}

// GetResult returns the discovery result that the Balancer holds.
func (bl *Balancer) GetResult() (res discovery.Result, ok bool) {
	if v := bl.res.Load(); v != nil {
		return v.(discovery.Result), true
	}
	return
}

// GetPicker equal to loadbalance.Balancer without pass discovery.Result, because we cache the result
func (bl *Balancer) GetPicker() loadbalance.Picker {
	atomic.StoreInt32(&bl.expire, 0)
	res := bl.res.Load().(discovery.Result)
	return bl.b.balancer.GetPicker(res)
}

func (bl *Balancer) close() {
	// notice the under rebalancer
	if rb, ok := bl.b.balancer.(loadbalance.Rebalancer); ok {
		// notice to rebalancing
		rb.Delete(discovery.Change{
			Result: discovery.Result{
				Cacheable: true,
				CacheKey:  bl.res.Load().(discovery.Result).CacheKey,
			},
		})
	}
	// delete from sharedTicker
	bl.sharedTicker.Delete(bl)
}

const unknown = "unknown"

func Dump() interface{} {
	type instInfo struct {
		Address string
		Weight  int
	}
	cacheDump := make(map[string]interface{})
	balancerFactories.Range(func(key, val interface{}) bool {
		cacheKey := key.(string)
		if bf, ok := val.(*BalancerFactory); ok {
			routeMap := make(map[string]interface{})
			cacheDump[cacheKey] = routeMap
			bf.cache.Range(func(k, v interface{}) bool {
				routeKey := k.(string)
				if bl, ok := v.(*Balancer); ok {
					if dr, ok := bl.res.Load().(discovery.Result); ok {
						insts := make([]instInfo, 0, len(dr.Instances))
						for i := range dr.Instances {
							inst := dr.Instances[i]
							addr := fmt.Sprintf("%s://%s", inst.Address().Network(), inst.Address().String())
							insts = append(insts, instInfo{Address: addr, Weight: inst.Weight()})
						}
						routeMap[routeKey] = insts
					} else {
						routeMap[routeKey] = unknown
					}
				} else {
					routeMap[routeKey] = unknown
				}
				return true
			})
		} else {
			cacheDump[cacheKey] = unknown
		}
		return true
	})
	return cacheDump
}
