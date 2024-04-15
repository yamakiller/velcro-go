package messages

import (
	"context"
	"encoding/binary"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/rpc/protocol"
	"github.com/yamakiller/velcro-go/utils"
)

const(
	AlignBit = uint32(4)
)
func Marshal(msgBytes []byte) ([]byte, error) {

	dataLen := len(msgBytes)

	packetLen := RpcHeaderLength + dataLen


	buffer := make([]byte, utils.AlignOf(uint32(packetLen), AlignBit))
	//

	var offset int = RpcHeaderLength
	offset += copy(buffer[offset:], msgBytes)
	binary.BigEndian.PutUint16(buffer[0:RpcHeaderLength], uint16(dataLen))
	return buffer[:packetLen], nil
}

func MarshalTStruct(ctx context.Context,iprot protocol.IProtocol, msg thrift.TStruct, name string,seqid int32) ([]byte,error){
	iprot.Release()
	if err := iprot.WriteMessageBegin(ctx,name,thrift.EXCEPTION,seqid);err != nil{
		return nil,err
	}
	if err:= msg.Write(ctx,iprot);err!=nil{
		return nil, err
	}
	if err := iprot.WriteMessageEnd(ctx); err != nil{
		return nil,err
	}
	b := iprot.GetBytes()
	iprot.Release()
	return b,nil
}

