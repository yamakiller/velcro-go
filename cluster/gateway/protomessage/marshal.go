package protomessge

import (
	"encoding/binary"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/encryption"
)

const (
	AlignBit   = uint32(4)
	HeaderSize = 2
)

// 2 总长度
// 总长度 == 2 + MessageLength

func Marshal(msgBytes []byte, secret []byte) ([]byte, error) {

	dataLen := len(msgBytes)

	packetLen := HeaderSize + dataLen

	var buffer []byte

	if secret != nil {
		buffer = make([]byte, utils.AlignOf(uint32(packetLen), uint32(len(secret))))
	} else {
		buffer = make([]byte, utils.AlignOf(uint32(packetLen), AlignBit))
	}
	//

	var offset int = HeaderSize
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
