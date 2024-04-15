package server

import (
	"context"

	"github.com/yamakiller/velcro-go/network"
)

type RpcClientContext struct {
	sequenceID int32
	background context.Context
	context    network.Context
	message    interface{}
}

func (rcc *RpcClientContext) SequenceID() int32 {
	return rcc.sequenceID
}

func (rcc *RpcClientContext) Background() context.Context {
	return rcc.background
}

func (rcc *RpcClientContext) Context() network.Context {
	return rcc.context
}

func (rcc *RpcClientContext) Message() interface{} {
	return rcc.message
}



type ctxRpcClientContextKeyType struct{}

var ctxRpcClientContextKey ctxRpcClientContextKeyType

func NewCtxWithRpcClientContext(ctx context.Context, si *RpcClientContext) context.Context {
	if si != nil {
		return context.WithValue(ctx, ctxRpcClientContextKey, si)
	}
	return ctx
}

func FreeCtxWithRpcClientContext(ctx context.Context) context.Context {
	ci := GetRpcClientContext(ctx)
	if ci == nil {
		return ctx
	}
	ci = nil

	return context.WithValue(ctx, ctxRpcClientContextKey, ci)
}

func GetRpcClientContext(ctx context.Context) *RpcClientContext {
	if ri, ok := ctx.Value(ctxRpcClientContextKey).(*RpcClientContext); ok {
		return ri
	}
	return nil
}