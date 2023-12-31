package messages

import (
	"encoding/binary"
	"errors"

	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"google.golang.org/protobuf/proto"
)

func UnMarshalProtobuf(buffer circbuf.Buffer) (int, interface{}, error) {
	if buffer.Length() < RpcHeaderLength {
		return 0, nil, nil
	}

	direct, _ := buffer.GetByte()
	switch RpcDirect(direct) {
	case RpcRequest:
		return unmarshalRequestProtobuf(buffer)
	case RpcResponse:
		return unmarshalResponseProtobuf(buffer)
	/*case RpcMessage:
	return unmarshalMessageProtobuf(buffer)*/
	case RpcPing:
		return unmarshalMessagePing(buffer)
	default:
		return 0, nil, errors.New("unknown message")
	}
}

func readHeader(data []byte) *RpcHeader {
	return &RpcHeader{Direct: RpcDirect(data[0]), Length: binary.BigEndian.Uint16(data[1:3])}
}

func unmarshalRequestProtobuf(buffer circbuf.Buffer) (int, proto.Message, error) {

	headerBytes := make([]byte, utils.AlignOf(uint32(RpcHeaderLength), uint32(4)))
	buffer.Get(headerBytes[:RpcHeaderLength])

	requestHeader := readHeader(headerBytes)

	if buffer.Length() < int(requestHeader.Length)+RpcHeaderLength {
		return 0, nil, nil
	}

	buffer.Peek(RpcHeaderLength)

	message, err := unmarshalProtobufBody(buffer, requestHeader.Direct, int(requestHeader.Length))
	if err != nil {
		return 0, nil, err
	}

	return int(requestHeader.Length) + RpcHeaderLength, message, nil
}

func unmarshalResponseProtobuf(buffer circbuf.Buffer) (int, proto.Message, error) {
	headerBytes := make([]byte, utils.AlignOf(uint32(RpcHeaderLength), uint32(4)))
	buffer.Get(headerBytes[:RpcHeaderLength])

	respHeader := readHeader(headerBytes)

	if buffer.Length() < int(respHeader.Length)+RpcHeaderLength {
		return 0, nil, nil
	}
	buffer.Peek(RpcHeaderLength)
	message, err := unmarshalProtobufBody(buffer, respHeader.Direct, int(respHeader.Length))
	if err != nil {
		return 0, nil, err
	}

	return int(respHeader.Length) + RpcHeaderLength, message, nil
}

func unmarshalMessageProtobuf(buffer circbuf.Buffer) (int, proto.Message, error) {
	headerBytes := make([]byte, utils.AlignOf(uint32(RpcHeaderLength), uint32(4)))
	buffer.Get(headerBytes[:RpcHeaderLength])

	respHeader := readHeader(headerBytes)

	if buffer.Length() < int(respHeader.Length)+RpcHeaderLength {
		return 0, nil, nil
	}

	buffer.Peek(RpcHeaderLength)
	message, err := unmarshalProtobufBody(buffer, respHeader.Direct, int(respHeader.Length))
	if err != nil {
		return 0, nil, err
	}

	return int(respHeader.Length) + RpcHeaderLength, message, nil
}

func unmarshalMessagePing(buffer circbuf.Buffer) (int, proto.Message, error) {
	headerBytes := make([]byte, utils.AlignOf(uint32(RpcHeaderLength), uint32(4)))
	buffer.Get(headerBytes[:RpcHeaderLength])

	respHeader := readHeader(headerBytes)

	if buffer.Length() < int(respHeader.Length)+RpcHeaderLength {
		return 0, nil, nil
	}

	buffer.Peek(RpcHeaderLength)
	message, err := unmarshalProtobufBody(buffer, respHeader.Direct, int(respHeader.Length))
	if err != nil {
		return 0, nil, err
	}

	return int(respHeader.Length) + RpcHeaderLength, message, nil
}

func unmarshalProtobufBody(buffer circbuf.Buffer, direct RpcDirect, bodyLength int) (proto.Message, error) {
	bodyBytes := make([]byte, utils.AlignOf(uint32(bodyLength), uint32(4)))
	buffer.Read(bodyBytes[:bodyLength])

	var msg proto.Message
	switch direct {
	case RpcRequest:
		msg = &RpcRequestMessage{}
	case RpcResponse:
		msg = &RpcResponseMessage{}
	/*case RpcMessage:
	msg = &RpcMsgMessage{}*/
	case RpcPing:
		msg = &RpcPingMessage{}
	}

	err := proto.Unmarshal(bodyBytes[:bodyLength], msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
