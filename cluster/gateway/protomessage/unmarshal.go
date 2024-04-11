package protomessge

import (
	"encoding/binary"


	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/encryption"
)

func UnMarshal(reader circbuf.Reader, secret []byte) ([]byte, error) {

	if reader.Len() < HeaderSize {
		return  nil, nil
	}

	HeaderByte, err := reader.Peek(HeaderSize)
	if err != nil {
		return nil, err
	}
	
	readDataLen := binary.BigEndian.Uint16(HeaderByte[:])
	if readDataLen+HeaderSize > uint16(reader.Len()) {
		return nil, nil
	}

	reader.Skip(HeaderSize)
	bodyByte, err := reader.ReadBinary(int(readDataLen))
	if err != nil {
		return nil, err
	}

	if secret != nil {
		decrypt, err := encryption.AesDecryptByGCM(bodyByte, secret)
		if err != nil {
			return nil, err
		}
		bodyByte = decrypt
	}
	return bodyByte, nil
}
