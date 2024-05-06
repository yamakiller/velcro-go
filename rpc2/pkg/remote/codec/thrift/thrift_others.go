package thrift

import "github.com/yamakiller/velcro-go/rpc2/pkg/remote"

// hyperMarshalEnabled indicates that if there are high priority message codec for current platform.
func (c thriftCodec) hyperMarshalEnabled() bool {
	return false
}

// hyperMarshalAvailable indicates that if high priority message codec is available.
func hyperMarshalAvailable(data interface{}) bool {
	return false
}

// hyperMessageUnmarshalEnabled indicates that if there are high priority message codec for current platform.
func (c thriftCodec) hyperMessageUnmarshalEnabled() bool {
	return false
}

// hyperMessageUnmarshalAvailable indicates that if high priority message codec is available.
func hyperMessageUnmarshalAvailable(data interface{}, payloadLen int) bool {
	return false
}

func (c thriftCodec) hyperMarshal(out remote.ByteBuffer, methodName string, msgType remote.MessageType, seqID int32, data interface{}) error {
	panic("unreachable code")
}

func (c thriftCodec) hyperMarshalBody(data interface{}) (buf []byte, err error) {
	panic("unreachable code")
}

func (c thriftCodec) hyperMessageUnmarshal(buf []byte, data interface{}) error {
	panic("unreachable code")
}
