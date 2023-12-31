package messages

type RpcDirect int
type RpcQos int

const (
	RpcRequest = iota
	RpcResponse
	RpcPing
)

const (
	RpcHeaderLength int = 3
)

type RpcHeader struct {
	Direct RpcDirect
	Length uint16
}

/*const (
	RpcRequestHeaderLength  int = 25
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
}*/

/*type RpcRequestMessage struct {
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
	Qos        RpcQos
	Message    interface{}
}

type RpcPingMessage struct {
	VerifyKey uint64
}*/
