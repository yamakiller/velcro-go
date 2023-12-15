package protomessge

import (
	"encoding/binary"
	"fmt"

	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/encryption"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func UnMarshal(in *circbuf.RingBuffer, secret []byte) (proto.Message, error) {

	if in.Length() < HeaderSize {
		return nil, nil
	}

	var HeaderByte [HeaderSize]byte
	in.Get(HeaderByte[:])

	readDataLen := binary.BigEndian.Uint16(HeaderByte[:])
	if readDataLen+HeaderSize < uint16(in.Length()) {
		return nil, nil
	}

	if readDataLen > uint16(float64(in.Capacity())*0.5) {
		return nil, fmt.Errorf("message data overflow %d/%d", readDataLen, uint16(float64(in.Capacity())*0.5))
	}

	in.Read(HeaderByte[:])
	b := make([]byte, utils.AlignOf(uint32(readDataLen), AlignBit))
	in.Read(b)

	dataLen := int(readDataLen)
	if secret != nil {
		decrypt, err := encryption.AesDecryptByGCM(b, secret)
		if err != nil {
			return nil, err
		}
		dataLen = copy(b[:len(decrypt)], decrypt)
	}

	//1.解析名称名称
	msgNameLen := int(b[0])
	if msgNameLen > (dataLen - 1) {
		return nil, fmt.Errorf("message name length error %d", msgNameLen)
	}

	proName := string(b[1 : msgNameLen+1])
	//2.解析Protobuf
	msgName := protoreflect.FullName(proName)
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(msgName)
	if err != nil {
		return nil, err
	}

	message := msgType.New().Interface()
	err = proto.Unmarshal(b[1+msgNameLen:dataLen], message)
	if err != nil {
		return nil, err
	}

	return message, nil
}
