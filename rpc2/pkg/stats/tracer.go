package stats

import "context"

// 跟踪在 RPC 开始和结束时执行
type Tracer interface {
	Start(ctx context.Context) context.Context
	Finish(ctx context.Context)
}
