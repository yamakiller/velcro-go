package messages

type RpcDirect int

const (
	RpcRequest = iota
	RpcResponse
	RpcMessage
	RpcPing
)

const (
	RpcRequestHeaderLength  int = 23
	RpcResponseHeaderLength int = 8
	RpcMessageHeaderLength  int = 7
	RpcPingMessageLength    int = 8
)

type RpcRequestHeader struct {
	SequenceID  int32
	ForwardTime uint64
	Timeout     uint64
	BodyLength  uint16
}

type RpcResonseHeader struct {
	SequenceID int32
	Result     int8
	BodyLength uint16
}

type RpcMessageHeader struct {
	SequenceID int32
	BodyLength uint16
}

type RpcRequestMessage struct {
	SequenceID  int32
	ForwardTime uint64
	Timeout     uint64
	Message     interface{}
}

type RpcResponseMessage struct {
	SequenceID int32
	Result     int8
	Message    interface{}
}

type RpcMsgMessage struct {
	SequenceID int32
	Message    interface{}
}

type RpcPingMessage struct {
	VerifyKey uint64
}
