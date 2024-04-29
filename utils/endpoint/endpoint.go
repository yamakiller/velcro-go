package endpoint

import "context"

// Endpoint 代表一个远程调用的方法.
type Endpoint func(ctx context.Context, req, resp interface{}) (err error)

// Middleware 具有输入端点和输出端点.
type Middleware func(Endpoint) Endpoint

// MiddlewareBuilder 使用上下文中的信息构建中间件.
type MiddlewareBuilder func(ctx context.Context) Middleware

// Chain 将中间件连接成一个中间件.
func Chain(mws ...Middleware) Middleware {
	return func(next Endpoint) Endpoint {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}

// Build 将给定的中间件构建为一个中间件.
func Build(mws []Middleware) Middleware {
	if len(mws) == 0 {
		return DummyMiddleware
	}
	return func(next Endpoint) Endpoint {
		return mws[0](Build(mws[1:])(next))
	}
}

// DummyMiddleware 是一个虚拟中间件.
func DummyMiddleware(next Endpoint) Endpoint {
	return next
}

// DummyEndpoint 是一个虚拟端点.
func DummyEndpoint(ctx context.Context, req, resp interface{}) (err error) {
	return nil
}

type mwCtxKeyType int

// 附加到中间件构建器上下文的组件的键
const (
	CtxEventBusKey mwCtxKeyType = iota
	CtxEventQueueKey
)
