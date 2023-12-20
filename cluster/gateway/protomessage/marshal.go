package protomessge

import (
	"encoding/binary"
	"strings"

	"github.com/yamakiller/velcro-go/utils/encryption"
	"google.golang.org/protobuf/proto"
)

const (
	AlignBit   = uint32(4)
	HeaderSize = 2
)

func Marshal(buffer []byte, message proto.Message, secret []byte) (int32, error) {
	msgBytes, err := proto.Marshal(message)
	if err != nil {
		return -1, err
	}

	msgName := proto.MessageName(message)
	msgNameLen := strings.Count(string(msgName), "")
	dataLen := msgNameLen + len(msgBytes) + 1

	var offset int = HeaderSize
	buffer[offset] = uint8(msgNameLen)
	offset++
	copy(buffer[offset:offset+msgNameLen], []byte(string(msgName)))
	offset += msgNameLen
	offset += copy(buffer[offset+msgNameLen:], msgBytes)
	if secret != nil {
		ebys, err := encryption.AesEncryptByGCM(buffer[HeaderSize:dataLen+HeaderSize], secret)
		if err != nil {
			return -1, err
		}
		binary.BigEndian.PutUint16(buffer[0:HeaderSize], uint16(len(ebys)))
		copy(buffer[HeaderSize:len(ebys)+HeaderSize], ebys)
		dataLen = len(ebys) + HeaderSize
	} else {
		dataLen += HeaderSize
	}

	return int32(dataLen), nil
}
