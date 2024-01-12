package messages

import (
	"encoding/binary"
	"time"

	"github.com/yamakiller/velcro-go/utils"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func MarshalRequestProtobuf(sequenceID int32, timeout uint64, message proto.Message) ([]byte, error) {
	msgAny, err := anypb.New(message)
	if err != nil {
		return nil, err
	}

	request := &RpcRequestMessage{
		SequenceID:  sequenceID,
		ForwardTime: uint64(time.Now().UnixMilli()),
		Timeout:     timeout,
		Message:     msgAny,
	}

	msgBytes, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	data := make([]byte, utils.AlignOf(uint32(len(msgBytes)+RpcHeaderLength), uint32(4)))
	data[0] = RpcRequest
	binary.BigEndian.PutUint16(data[1:3], uint16(len(msgBytes)))
	n := copy(data[RpcHeaderLength:len(msgBytes)+RpcHeaderLength], msgBytes)

	return data[:n+RpcHeaderLength], nil
}

func MarshalResponseProtobuf(sequenceID int32, result proto.Message) ([]byte, error) {
	var (
		resultAny *anypb.Any
		err error
	)
	if result != nil{
		resultAny, err = anypb.New(result)
		if err != nil {
			return nil, err
		}
	}

	resp := &RpcResponseMessage{
		SequenceID: sequenceID,
		Result:     resultAny,
	}

	respBytes, err := proto.Marshal(resp)
	if err != nil {
		return nil, err
	}

	data := make([]byte, utils.AlignOf(uint32(len(respBytes)+RpcHeaderLength), uint32(4)))
	data[0] = RpcResponse
	binary.BigEndian.PutUint16(data[1:3], uint16(len(respBytes)))
	n := copy(data[RpcHeaderLength:len(respBytes)+RpcHeaderLength], respBytes)

	return data[:n+RpcHeaderLength], nil
}

/*func MarshalMessageProtobuf(sequenceID int32, message proto.Message) ([]byte, error) {
	msgAny, err := anypb.New(message)
	if err != nil {
		return nil, err
	}

	msg := &RpcMsgMessage{
		SequenceID: sequenceID,
		Message:    msgAny,
	}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	data := make([]byte, utils.AlignOf(uint32(len(msgBytes)+3), uint32(4)))
	data[0] = RpcMessage

	binary.BigEndian.PutUint16(data[1:3], uint16(len(msgBytes)))
	n := copy(data[3:len(msgBytes)+3], msgBytes)

	return data[:n], nil
}*/

func MarshalPingProtobuf(VerifyKey uint64) ([]byte, error) {
	msg := &RpcPingMessage{
		VerifyKey: VerifyKey,
	}

	respBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	data := make([]byte, utils.AlignOf(uint32(len(respBytes)+RpcHeaderLength), uint32(4)))
	data[0] = RpcPing
	binary.BigEndian.PutUint16(data[1:3], uint16(len(respBytes)))
	n := copy(data[RpcHeaderLength:len(respBytes)+RpcHeaderLength], respBytes)

	return data[:n+RpcHeaderLength], nil
}
