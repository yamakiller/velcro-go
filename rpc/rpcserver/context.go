package rpcserver

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
