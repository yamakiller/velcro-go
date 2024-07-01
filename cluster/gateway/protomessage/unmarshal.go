package protomessge

import (
	"encoding/binary"
	"fmt"

	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/encryption"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func UnMarshal(reader circbuf.Reader, secret []byte) (proto.Message, error) {

	if reader.Len() < HeaderSize {
		return nil, nil
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

	if len(bodyByte) == 0 {
		return nil, fmt.Errorf("bodyByte length error %d", readDataLen)
	}
	dataLen := int(readDataLen)
	if secret != nil {
		decrypt, err := encryption.AesDecryptByGCM(bodyByte, secret)
		if err != nil {
			return nil, err
		}

		bodyByte = decrypt
		dataLen = len(decrypt)
	}

	//1.解析名称名称
	msgNameLen := int(bodyByte[0])
	if msgNameLen > (dataLen - 1) {
		return nil, fmt.Errorf("message name length error %d", msgNameLen)
	}

	if len(bodyByte) < msgNameLen+1 {
		return nil, fmt.Errorf("message name length error %d", len(bodyByte))
	}

	proName := string(bodyByte[1 : msgNameLen+1])
	//2.解析Protobuf
	msgName := protoreflect.FullName(proName)
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(msgName)
	if err != nil {
		return nil, fmt.Errorf("msgName %v  %v",msgName,err.Error())
	}

	message := msgType.New().Interface()
	err = proto.Unmarshal(bodyByte[1+msgNameLen:dataLen], message)
	if err != nil {
		return nil, err
	}

	return message, nil
}
