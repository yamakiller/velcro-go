package messages

import (
	"github.com/yamakiller/velcro-go/utils"
	"google.golang.org/protobuf/proto"
)

func MarshalRequestProtobuf(sequenceID int32, timeout uint64, message interface{}) ([]byte, error) {
	protomessage := message.(proto.Message)
	msgBytes, err := proto.Marshal(protomessage)
	if err != nil {
		return nil, err
	}

	msgName := proto.MessageName(protomessage)
	msgNameBytes := []byte(string(msgName))
	msgNameLen := len(msgNameBytes)

	var msgBodyBytes []byte = make([]byte, 0)

	msgBodyBytes = append(msgBodyBytes, uint8(msgNameLen))
	msgBodyBytes = append(msgBodyBytes, msgNameBytes[:]...)
	msgBodyBytes = append(msgBodyBytes, msgBytes[:]...)

	var msgBuffer []byte = make([]byte, utils.AlignOf(uint32(len(msgBodyBytes)+RpcRequestHeaderLength), uint32(4)))
	length, err := marshalRequest(msgBuffer[:len(msgBodyBytes)+RpcRequestHeaderLength], sequenceID, timeout, msgBodyBytes)
	if err != nil {
		return nil, err
	}
	return msgBuffer[:length], nil
}

func MarshalResponseProtobuf(sequenceID int32, result int8, message interface{}) ([]byte, error) {
	protomessage := message.(proto.Message)
	var msgBodyBytes []byte = nil
	var msgBodyLength int = 0
	if message != nil {
		msgBytes, err := proto.Marshal(protomessage)
		if err != nil {
			return nil, err
		}

		msgName := proto.MessageName(protomessage)
		msgNameBytes := []byte(string(msgName))
		msgNameLen := len(msgNameBytes)

		msgBodyBytes = make([]byte, 0)
		msgBodyBytes = append(msgBodyBytes, uint8(msgNameLen))
		msgBodyBytes = append(msgBodyBytes, msgNameBytes[:]...)
		msgBodyBytes = append(msgBodyBytes, msgBytes[:]...)

		msgBodyLength = len(msgBodyBytes)
	}

	var msgBuffer []byte = make([]byte, utils.AlignOf(uint32(msgBodyLength+RpcResponseHeaderLength), uint32(4)))
	length, err := marshalResponse(msgBuffer[:msgBodyLength+RpcResponseHeaderLength], sequenceID, result, msgBodyBytes)
	if err != nil {
		return nil, err
	}
	return msgBuffer[:length], nil
}

func MarshalMessageProtobuf(sequenceID int32, message interface{}) ([]byte, error) {
	protomessage := message.(proto.Message)
	msgBytes, err := proto.Marshal(protomessage)
	if err != nil {
		return nil, err
	}

	msgName := proto.MessageName(protomessage)
	msgNameBytes := []byte(string(msgName))
	msgNameLen := len(msgNameBytes)

	var msgBodyBytes []byte = make([]byte, 0)

	msgBodyBytes = append(msgBodyBytes, uint8(msgNameLen))
	msgBodyBytes = append(msgBodyBytes, msgNameBytes[:]...)
	msgBodyBytes = append(msgBodyBytes, msgBytes[:]...)

	var msgBodyLength int = len(msgBodyBytes)

	var msgBuffer []byte = make([]byte, utils.AlignOf(uint32(msgBodyLength+RpcMessageHeaderLength), uint32(4)))
	length, err := marshalMessage(msgBuffer[:msgBodyLength+RpcMessageHeaderLength], sequenceID, msgBodyBytes)
	if err != nil {
		return nil, err
	}
	return msgBuffer[:length], nil
}
