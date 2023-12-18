package rpc

import "errors"

var (
	ErrorRequestTimeout = errors.New("rpc request: timeout")
)
