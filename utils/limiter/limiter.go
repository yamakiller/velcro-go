package limiter

import (
	"context"
	"time"

	"github.com/yamakiller/velcro-go/cluster/utils/limit"
)

// ConcurrencyLimiter 限制对受保护资源的并发访问数量;
// 注意 ConcurrencyLimiter 的实现应该是并发安全的.
type ConcurrencyLimiter interface {
	// Acquire 报告是否允许下次访问受保护资源.
	Acquire(ctx context.Context) bool

	// Release 声称先前的访问已释放该资源.
	Release(ctx context.Context)

	// Status 返回总配额和已占用的配额.
	Status(ctx context.Context) (limit, occupied int)
}

// RateLimiter 限制对受保护资源的访问速率
type RateLimiter interface {
	// Acquire 报告是否允许下次访问受保护资源.
	Acquire(ctx context.Context) bool
	// Status 状态返回速率限制.
	Status(ctx context.Context) (max, current int, interval time.Duration)
}

// Updatable 是一种支持动态更改限制的限制器;
// 请注意, "UpdateLimit"不保证并发安全.
type Updatable interface {
	UpdateLimit(limit int)
}

// LimitReporter 是定义在发生限制时报告（指标或打印日志）的接口.
type LimitReporter interface {
	ConnOverloadReport()
	QPSOverloadReport()
}

// NewLimiterWrapper 创建一个限制代理;
// 将给定的 ConcurrencyLimiter 和 RateLimiter 包装到 limit.Updater 中.
func NewLimiterWrapper(conLimit ConcurrencyLimiter, qpsLimit RateLimiter) limit.Updater {
	return &limitWrapper{
		conLimit: conLimit,
		qpsLimit: qpsLimit,
	}
}

type limitWrapper struct {
	conLimit ConcurrencyLimiter
	qpsLimit RateLimiter
}

func (l *limitWrapper) UpdateLimit(opt *limit.Option) (updated bool) {
	if opt != nil {
		if ul, ok := l.conLimit.(Updatable); ok {
			ul.UpdateLimit(opt.MaxConnections)
			updated = true
		}
		if ul, ok := l.qpsLimit.(Updatable); ok {
			ul.UpdateLimit(opt.MaxQPS)
			updated = true
		}
	}
	return
}
