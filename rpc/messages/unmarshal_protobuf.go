package messages

import (
	"errors"

	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func UnMarshalProtobuf(buffer circbuf.Buffer) (int, interface{}, error) {
	if buffer.Length() < 1 {
		return 0, nil, nil
	}

	direct, _ := buffer.GetByte()
	switch RpcDirect(direct) {
	case RpcRequest:
		return unmarshalRequestProtobuf(buffer)
	case RpcResponse:
		return unmarshalResponseProtobuf(buffer)
	case RpcMessage:
		return unmarshalMessageProtobuf(buffer)
	case RpcPing:
		return unmarshalPing(buffer)
	default:
		return 0, nil, errors.New("unknown message")
	}

	return -1, nil, nil
}

func unmarshalRequestProtobuf(buffer circbuf.Buffer) (int, *RpcRequestMessage, error) {
	if buffer.Length() < RpcRequestHeaderLength {
		return 0, nil, nil
	}

	headerBytes := make([]byte, utils.AlignOf(uint32(RpcRequestHeaderLength), uint32(4)))
	buffer.Get(headerBytes[:RpcRequestHeaderLength])

	requestHeader := readRequestHeader(headerBytes)

	if buffer.Length() < int(requestHeader.BodyLength)+RpcRequestHeaderLength {
		return 0, nil, nil
	}

	buffer.Peek(RpcRequestHeaderLength)

	message, err := unmarshalProtobufBody(buffer, int(requestHeader.BodyLength))
	if err != nil {
		return 0, nil, err
	}

	request := &RpcRequestMessage{SequenceID: requestHeader.SequenceID,
		ForwardTime: requestHeader.ForwardTime,
		Timeout:     requestHeader.Timeout,
		Message:     message}

	return int(requestHeader.BodyLength) + RpcRequestHeaderLength, request, nil
}

func unmarshalResponseProtobuf(buffer circbuf.Buffer) (int, *RpcResponseMessage, error) {
	if buffer.Length() < RpcResponseHeaderLength {
		return 0, nil, nil
	}

	headerBytes := make([]byte, utils.AlignOf(uint32(RpcResponseHeaderLength), uint32(4)))
	buffer.Get(headerBytes[:RpcResponseHeaderLength])

	respHeader := readResponseHeader(headerBytes)

	if buffer.Length() < int(respHeader.BodyLength)+RpcResponseHeaderLength {
		return 0, nil, nil
	}
	buffer.Peek(RpcResponseHeaderLength)
	message, err := unmarshalProtobufBody(buffer, int(respHeader.BodyLength))
	if err != nil {
		return 0, nil, err
	}

	return int(respHeader.BodyLength) + RpcResponseHeaderLength,
		&RpcResponseMessage{SequenceID: respHeader.SequenceID, Result: respHeader.Result, Message: message},
		nil
}

func unmarshalMessageProtobuf(buffer circbuf.Buffer) (int, *RpcMsgMessage, error) {
	if buffer.Length() < RpcMessageHeaderLength {
		return 0, nil, nil
	}

	headerBytes := make([]byte, utils.AlignOf(uint32(RpcMessageHeaderLength), uint32(4)))
	buffer.Get(headerBytes[:RpcMessageHeaderLength])
	msgHeader := readMessageHeader(headerBytes[:RpcMessageHeaderLength])
	if buffer.Length() < int(msgHeader.BodyLength)+RpcMessageHeaderLength {
		return 0, nil, nil
	}

	buffer.Peek(RpcMessageHeaderLength)
	message, err := unmarshalProtobufBody(buffer, int(msgHeader.BodyLength))
	if err != nil {
		return 0, nil, err
	}

	return int(msgHeader.BodyLength) + RpcMessageHeaderLength,
		&RpcMsgMessage{SequenceID: msgHeader.SequenceID, Message: message},
		nil
}

func unmarshalProtobufBody(buffer circbuf.Buffer, bodyLength int) (proto.Message, error) {
	bodyBytes := make([]byte, utils.AlignOf(uint32(bodyLength), uint32(4)))
	buffer.Read(bodyBytes[:bodyLength])

	offset := 0
	msgNameLen := int(bodyBytes[offset])
	offset++
	msgNameName := string(bodyBytes[offset : offset+msgNameLen])
	offset += msgNameLen

	msgType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(msgNameName))
	if err != nil {
		return nil, err
	}

	msgLen := int(bodyLength) - offset
	message := msgType.New().Interface()
	err = proto.Unmarshal(bodyBytes[offset:offset+msgLen], message)
	if err != nil {
		return nil, err
	}

	return message, nil
}
