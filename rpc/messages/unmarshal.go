package messages

import (
	"encoding/binary"

	"github.com/yamakiller/velcro-go/utils/circbuf"
)

type UnMarshalFunc func(buffer circbuf.Buffer) (int, interface{}, error)

func unmarshalPing(buffer circbuf.Buffer) (int, *RpcPingMessage, error) {
	if buffer.Length() < RpcPingMessageLength {
		return 0, nil, nil
	}
	var data [8]byte
	buffer.ReadByte()
	buffer.Read(data[:])

	pingMsg := &RpcPingMessage{
		VerifyKey: binary.BigEndian.Uint64(data[:]),
	}

	return RpcPingMessageLength, pingMsg, nil
}

func readRequestHeader(data []byte) *RpcRequestHeader {

	request := &RpcRequestHeader{}
	offset := 1

	request.SequenceID = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4

	request.ForwardTime = binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	request.Timeout = binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	request.BodyLength = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	return request
}

func readResponseHeader(data []byte) *RpcResonseHeader {
	resp := &RpcResonseHeader{}
	offset := 1

	resp.SequenceID = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4

	resp.Result = int8(data[offset : offset+1][0])
	offset += 1

	resp.BodyLength = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	return resp
}

func readMessageHeader(data []byte) *RpcMessageHeader {
	msg := &RpcMessageHeader{}
	offset := 1

	msg.SequenceID = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4

	msg.BodyLength = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	return msg
}
