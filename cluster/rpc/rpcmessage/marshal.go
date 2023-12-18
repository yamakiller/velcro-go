package rpcmessage

import (
	"encoding/binary"
	"errors"
	"time"
)

type MarshalRequestFunc func(sequenceID int32, timeout uint64, message interface{}) ([]byte, error)
type MarshalResponseFunc func(sequenceID int32, result int8, message interface{}) ([]byte, error)
type MarshalMessageFunc func(sequenceID int32, message interface{}) ([]byte, error)
type MarshalPingFunc func(value uint64) ([]byte, error)

// MarshalRequest 构建 Request 消息
func marshalRequest(buffer []byte, sequenceID int32, timeout uint64, message []byte /*包括消息名*/) (int, error) {
	//1.计算总长度
	if len(buffer) < (RpcRequestHeaderLength + len(message)) {
		return 0, errors.New("buffer: overflow")
	}

	// 数据包类型
	offset := 0
	buffer[offset] = uint8(RpcRequest)
	offset++
	// 包序列号
	binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(sequenceID))
	offset += 4
	// 发送时间
	binary.BigEndian.PutUint64(buffer[offset:offset+8], uint64(time.Now().UnixMilli()))
	offset += 8
	// 超时时间
	binary.BigEndian.PutUint64(buffer[offset:offset+8], timeout)
	offset += 8
	// 数据体长度
	binary.BigEndian.PutUint16(buffer[offset:offset+2], uint16(len(message)))
	offset += 2
	// 数据体
	copy(buffer[RpcRequestHeaderLength:RpcRequestHeaderLength+len(message)], message)
	offset += len(message)
	return offset, nil
}

func marshalResponse(buffer []byte, sequenceID int32, result int8, message []byte /*包括消息名, 可为空*/) (int, error) {
	messageLength := 0
	if message != nil {
		messageLength = len(message)
	}
	if len(buffer) < (RpcResponseHeaderLength + messageLength) {
		return 0, errors.New("buffer: overflow")
	}

	// 数据包类型
	offset := 0
	buffer[offset] = uint8(RpcResponse)
	offset++

	// 包序列号
	binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(sequenceID))
	offset += 4
	// 结果
	buffer[offset] = byte(result)
	offset++
	// 数据体长度
	binary.BigEndian.PutUint16(buffer[offset:offset+2], uint16(messageLength))
	offset += 2

	if message != nil {
		copy(buffer[RpcResponseHeaderLength:RpcResponseHeaderLength+messageLength], message)
		offset += messageLength
	}

	return offset, nil
}

func marshalMessage(buffer []byte, sequenceID int32, message []byte) (int, error) {
	if len(buffer) < (RpcMessageHeaderLength + len(message)) {
		return 0, errors.New("buffer: overflow")
	}

	// 数据包类型
	offset := 0
	buffer[offset] = uint8(RpcResponse)
	offset++

	// 包序列号
	binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(sequenceID))
	offset += 4

	binary.BigEndian.PutUint16(buffer[offset:offset+2], uint16(len(message)))
	offset += 2

	copy(buffer[RpcMessageHeaderLength:RpcMessageHeaderLength+len(message)], message)
	offset += len(message)

	return offset, nil
}

func marshalPing(VerifyKey uint64) ([]byte, error) {
	var data [9]byte
	data[0] = RpcPing
	binary.BigEndian.PutUint64(data[1:], VerifyKey)

	return data[:], nil
}
