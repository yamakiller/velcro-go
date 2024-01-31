package protomessge

import (
	"encoding/binary"

	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/encryption"
	"google.golang.org/protobuf/proto"
)

const (
	AlignBit   = uint32(4)
	HeaderSize = 2
)

// 2 总长度
// 总长度 == 1 + MessageNameLength + MessageLength

func Marshal(message proto.Message, secret []byte) ([]byte, error) {
	msgBytes, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}

	msgName := proto.MessageName(message)
	msgNameLen :=len(msgName) //strings.Count(string(msgName), "") - 1
	dataLen := msgNameLen + len(msgBytes) + 1

	packetLen := HeaderSize + dataLen

	var buffer []byte

	if secret != nil {
		buffer = make([]byte, utils.AlignOf(uint32(packetLen), uint32(len(secret))))
	} else {
		buffer = make([]byte, utils.AlignOf(uint32(packetLen), AlignBit))
	}
	//

	var offset int = HeaderSize
	buffer[offset] = uint8(msgNameLen)
	offset++

	copy(buffer[offset:offset+msgNameLen], []byte(string(msgName)))
	offset += msgNameLen
	offset += copy(buffer[offset:], msgBytes)
	if secret != nil {
		ebys, err := encryption.AesEncryptByGCM(buffer[HeaderSize:dataLen+HeaderSize], secret)
		if err != nil {
			return nil, err
		}

		dataLen = len(ebys)
		packetLen = len(ebys) + HeaderSize
		if packetLen > len(buffer) {
			// 需要重新分配
			buffer = make([]byte, utils.AlignOf(uint32(packetLen), AlignBit))
		}
		binary.BigEndian.PutUint16(buffer[0:HeaderSize], uint16(dataLen))
		copy(buffer[HeaderSize:dataLen], ebys)
	}else{
		binary.BigEndian.PutUint16(buffer[0:HeaderSize], uint16(dataLen))
	}

	return buffer[:packetLen], nil
}
