package limiter

import (
	"context"
	"sync/atomic"
	"time"
)

var fixedWindowTime = time.Second

type qpsLimiter struct {
	limit      int32
	tokens     int32
	interval   time.Duration
	once       int32
	ticker     *time.Ticker
	tickerDone chan bool
}

// UpdateLimit 更新间隔和限制;
// 它不是并发安全的
func (l *qpsLimiter) UpdateLimit(limit int) {
	once := calcOnce(l.interval, limit)
	atomic.StoreInt32(&l.limit, int32(limit))
	atomic.StoreInt32(&l.once, once)
	l.resetTokens(once)
}

// UpdateQPSLimit update the interval and limit.;
// 它不是并发安全的.
func (l *qpsLimiter) UpdateQPSLimit(interval time.Duration, limit int) {
	once := calcOnce(interval, limit)
	atomic.StoreInt32(&l.limit, int32(limit))
	atomic.StoreInt32(&l.once, once)
	l.resetTokens(once)
	if interval != l.interval {
		l.interval = interval
		l.stopTicker()
		go l.startTicker(interval)
	}
}

// Acquire one token.
func (l *qpsLimiter) Acquire(ctx context.Context) bool {
	if atomic.LoadInt32(&l.limit) <= 0 {
		return true
	}
	if atomic.LoadInt32(&l.tokens) <= 0 {
		return false
	}
	return atomic.AddInt32(&l.tokens, -1) >= 0
}

func (l *qpsLimiter) Status(ctx context.Context) (max, cur int, interval time.Duration) {
	max = int(atomic.LoadInt32(&l.limit))
	cur = int(atomic.LoadInt32(&l.tokens))
	interval = l.interval
	return
}

func (l *qpsLimiter) startTicker(interval time.Duration) {
	l.ticker = time.NewTicker(interval)
	defer l.ticker.Stop()
	l.tickerDone = make(chan bool, 1)
	tc := l.ticker.C
	td := l.tickerDone
	// ticker 和tickerDone 可以重置，不能直接使用l.ticker 或l.tickerDone
	for {
		select {
		case <-tc:
			l.updateToken()
		case <-td:
			return
		}
	}
}

func (l *qpsLimiter) stopTicker() {
	if l.tickerDone == nil {
		return
	}
	select {
	case l.tickerDone <- true:
	default:
	}
}

// 这里允许一些偏差以获得更好的性能.
func (l *qpsLimiter) updateToken() {
	if atomic.LoadInt32(&l.limit) < atomic.LoadInt32(&l.tokens) {
		return
	}

	once := atomic.LoadInt32(&l.once)

	delta := atomic.LoadInt32(&l.limit) - atomic.LoadInt32(&l.tokens)

	if delta > once || delta < 0 {
		delta = once
	}

	newTokens := atomic.AddInt32(&l.tokens, delta)
	if newTokens < once {
		atomic.StoreInt32(&l.tokens, once)
	}
}

func (l *qpsLimiter) resetTokens(once int32) {
	if atomic.LoadInt32(&l.tokens) > once {
		atomic.StoreInt32(&l.tokens, once)
	}
}

func calcOnce(interval time.Duration, limit int) int32 {
	if interval > time.Second {
		interval = time.Second
	}
	once := int32(float64(limit) / (fixedWindowTime.Seconds() / interval.Seconds()))
	if once < 0 {
		once = 0
	}
	return once
}
