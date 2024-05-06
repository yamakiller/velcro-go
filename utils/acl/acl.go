package acl

import (
	"context"
	"errors"

	"github.com/yamakiller/velcro-go/utils/endpoint"
	"github.com/yamakiller/velcro-go/utils/verrors"
)

type RejectFunc func(ctx context.Context, request interface{}) (reason error)

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
