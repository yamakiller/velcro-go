package utils

// VelcroArgs is used for assert when get real request from XXXArgs.
// Thrift and VelcroProtobuf will generate GetFirstArgument() interface{} for XXXArgs
type VelcroArgs interface {
	GetFirstArgument() interface{}
}

// VelcroResult is used for assert when get real response from XXXResult.
// Thrift and VelcroProtobuf will generate the two functions for XXXResult.
type VelcroResult interface {
	GetResult() interface{}
	SetSuccess(interface{})
}
