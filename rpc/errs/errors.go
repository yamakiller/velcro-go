package errs

import "errors"

var (
	ErrorRpcConnectorClosed         = errors.New("rpc connector: closed")
	ErrorRequestTimeout             = errors.New("rpc request: timeout")
	ErrorRpcUnMarshal               = errors.New("rpc request: unmarshal fail")
	ErrorRpcConnectPoolNoInitialize = errors.New("rpc connect pool: no initial")
	ErrorRpcConnectMaxmize          = errors.New("rpc connect is maximize")
	ErrorInvalidRpcConnect          = errors.New("invalid rpc connect")
	ErrorRpcConnPoolClosed          = errors.New("rpc connector pool: closed")
)
