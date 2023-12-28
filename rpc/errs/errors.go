package errs

import "errors"

var (
	ErrorRpcConnectorClosed = errors.New("rpc connector: closed")
	ErrorRequestTimeout     = errors.New("rpc request: timeout")
	ErrorRpcUnMarshal       = errors.New("rpc request: unmarshal fail")
)
