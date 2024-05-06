package lbcache

import (
	"sync"
	"time"

	"github.com/yamakiller/velcro-go/utils"
	"golang.org/x/sync/singleflight"
)

var (
	// insert, not delete
	sharedTickers    sync.Map
	sharedTickersSfg singleflight.Group
)

func getSharedTicker(b *Balancer, refreshInterval time.Duration) *utils.SharedTicker {
	sti, ok := sharedTickers.Load(refreshInterval)
	if ok {
		st := sti.(*utils.SharedTicker)
		st.Add(b)
		return st
	}

	v, _, _ := sharedTickersSfg.Do(refreshInterval.String(), func() (interface{}, error) {
		st := utils.NewSharedTicker(refreshInterval)
		sharedTickers.Store(refreshInterval, st)
		return st, nil
	})
	st := v.(*utils.SharedTicker)
	// Add without singleflight,
	// 因为我们需要所有调用此函数的复习者将自己添加到 SharedTicker
	st.Add(b)
	return st
}
