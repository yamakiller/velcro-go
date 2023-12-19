package rpc

import "errors"

var (
	ErrorRequestTimeout  = errors.New("rpc request: timeout")
	ErrorRpcClientClosed = errors.New("rpc request: client closed")
)
