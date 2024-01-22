package acl

import (
	"context"
	"errors"

	"github.com/yamakiller/velcro-go/rpc/utils/endpoint"
	"github.com/yamakiller/velcro-go/rpc/utils/verrors"
)

// RejectFunc 根据给定的上下文和请求判断是否拒绝请求.
// 如果被拒绝则返回原因，否则返回 nil.
type RejectFunc func(ctx context.Context, request interface{}) (reason error)

// NewACLMiddleware creates a new ACL middleware using the provided reject funcs.
func NewACLMiddleware(rules []RejectFunc) endpoint.Middleware {
	if len(rules) == 0 {
		return endpoint.DummyMiddleware
	}
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request, response interface{}) error {
			for _, r := range rules {
				if e := r(ctx, request); e != nil {
					if !errors.Is(e, verrors.ErrACL) {
						e = verrors.ErrACL.WithCause(e)
					}
					return e
				}
			}
			return next(ctx, request, response)
		}
	}
}
