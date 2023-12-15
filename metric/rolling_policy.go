package metric

import (
	"sync"
	"time"
)

// RollingPolicy 是基于持续时间的环窗策略.
// RollingPolicy 随着时间的推移移动铲斗偏移量.
// e.g. 如果最后一个点是在一个桶持续时间之前附加的，
// RollingPolicy 将增加当前偏移量.
type RollingPolicy struct {
	mu     sync.RWMutex
	size   int
	window *Window
	offset int

	bucketDuration time.Duration
	lastAppendTime time.Time
}

// RollingPolicyOpts 包含创建 RollingPolicy 的参数.
type RollingPolicyOpts struct {
	BucketDuration time.Duration
}

// NewRollingPolicy 根据给定的窗口和 RollingPolicyOpts 创建一个新的 RollingPolicy.
func NewRollingPolicy(window *Window, opts RollingPolicyOpts) *RollingPolicy {
	return &RollingPolicy{
		window: window,
		size:   window.Size(),
		offset: 0,

		bucketDuration: opts.BucketDuration,
		lastAppendTime: time.Now(),
	}
}

func (r *RollingPolicy) timespan() int {
	v := int(time.Since(r.lastAppendTime) / r.bucketDuration)
	if v > -1 { // maybe time backwards
		return v
	}
	return r.size
}

func (r *RollingPolicy) add(f func(offset int, val float64), val float64) {
	r.mu.Lock()
	timespan := r.timespan()
	if timespan > 0 {
		r.lastAppendTime = r.lastAppendTime.Add(time.Duration(timespan * int(r.bucketDuration)))
		offset := r.offset
		// 重置过期的桶
		s := offset + 1
		if timespan > r.size {
			timespan = r.size
		}
		e, e1 := s+timespan, 0 // e: 重置偏移量必须从 offset+1 开始
		if e > r.size {
			e1 = e - r.size
			e = r.size
		}
		for i := s; i < e; i++ {
			r.window.ResetBucket(i)
			offset = i
		}
		for i := 0; i < e1; i++ {
			r.window.ResetBucket(i)
			offset = i
		}
		r.offset = offset
	}
	f(r.offset, val)
	r.mu.Unlock()
}

// Append 将给定点附加到窗口.
func (r *RollingPolicy) Append(val float64) {
	r.add(r.window.Append, val)
}

// Add 将给定值添加到存储桶内的最新点.
func (r *RollingPolicy) Add(val float64) {
	r.add(r.window.Add, val)
}

// Reduce 将归约函数应用于窗口内的所有桶.
func (r *RollingPolicy) Reduce(f func(Iterator) float64) (val float64) {
	r.mu.RLock()
	timespan := r.timespan()
	if count := r.size - timespan; count > 0 {
		offset := r.offset + timespan + 1
		if offset >= r.size {
			offset = offset - r.size
		}
		val = f(r.window.Iterator(offset, count))
	}
	r.mu.RUnlock()
	return val
}
