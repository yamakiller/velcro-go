package messages

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
)

func UnMarshalProtobuf(reader circbuf.Reader) (int, interface{}, error) {
	if reader.Len() < RpcHeaderLength {
		return 0, nil, nil
	}

	direct, _ := reader.Peek(1)
	switch RpcDirect(direct[0]) {
	case RpcRequest:
		return unmarshalRequestProtobuf(reader)
	case RpcResponse:
		return unmarshalResponseProtobuf(reader)
	/*case RpcMessage:
	return unmarshalMessageProtobuf(buffer)*/
	case RpcPing:
		return unmarshalMessagePing(reader)
	default:
		return 0, nil, errors.New("unknown message")
	}
}

func readHeader(data []byte) *RpcHeader {
	return &RpcHeader{Direct: RpcDirect(data[0]), Length: binary.BigEndian.Uint16(data[1:3])}
}

func unmarshalRequestProtobuf(reader circbuf.Reader) (int, proto.Message, error) {

	headerBytes, err := reader.Peek(RpcHeaderLength)
	if err != nil {
		return 0, nil, err
	}

	requestHeader := readHeader(headerBytes)

	if reader.Len() < int(requestHeader.Length)+RpcHeaderLength {
		return 0, nil, nil
	}

	reader.Skip(RpcHeaderLength)

	message, err := unmarshalProtobufBody(reader, requestHeader.Direct, int(requestHeader.Length))
	if err != nil {
		return 0, nil, err
	}

	return int(requestHeader.Length) + RpcHeaderLength, message, nil
}

func unmarshalResponseProtobuf(reader circbuf.Reader) (int, proto.Message, error) {
	headerBytes, err := reader.Peek(RpcHeaderLength)
	if err != nil {
		return 0, nil, err
	}

	respHeader := readHeader(headerBytes)

	if reader.Len() < int(respHeader.Length)+RpcHeaderLength {
		return 0, nil, nil
	}

	reader.Skip(RpcHeaderLength)

	message, err := unmarshalProtobufBody(reader, respHeader.Direct, int(respHeader.Length))
	if err != nil {
		return 0, nil, err
	}

	return int(respHeader.Length) + RpcHeaderLength, message, nil
}

func unmarshalMessagePing(reader circbuf.Reader) (int, proto.Message, error) {
	headerBytes, err := reader.Peek(RpcHeaderLength)
	if err != nil {
		return 0, nil, err
	}

	respHeader := readHeader(headerBytes)

	if reader.Len() < int(respHeader.Length)+RpcHeaderLength {
		return 0, nil, nil
	}

	reader.Skip(RpcHeaderLength)
	message, err := unmarshalProtobufBody(reader, respHeader.Direct, int(respHeader.Length))
	if err != nil {
		return 0, nil, err
	}

	return int(respHeader.Length) + RpcHeaderLength, message, nil
}

func unmarshalProtobufBody(reader circbuf.Reader, direct RpcDirect, bodyLength int) (proto.Message, error) {

	bodyBytes, err := reader.ReadBinary(bodyLength)
	if err != nil {
		return nil, err
	}

	var msg proto.Message
	switch direct {
	case RpcRequest:
		msg = &RpcRequestMessage{}
	case RpcResponse:
		msg = &RpcResponseMessage{}
	case RpcPing:
		msg = &RpcPingMessage{}
	}

	err = proto.Unmarshal(bodyBytes[:bodyLength], msg)
	if err != nil {
		return nil, err
	}
	if direct == RpcRequest && msg.(*RpcRequestMessage).Message == nil {
		vlog.Error(fmt.Sprintf("RpcRequestMessage bodyLength %v bodyBytes %v", bodyLength, bodyBytes))
		return nil, fmt.Errorf("RpcRequestMessage bodyLength %v bodyBytes %v", bodyLength, bodyBytes)
	}
	return msg, nil
}
