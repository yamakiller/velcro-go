package messages

import (
	"context"
	"encoding/binary"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/rpc/protocol"
	"github.com/yamakiller/velcro-go/utils/circbuf"
)

func UnMarshal(reader circbuf.Reader) ([]byte, error) {
	if reader.Length() < RpcHeaderLength {
		return nil, nil
	}

	HeaderByte, err := reader.Peek(RpcHeaderLength)
	if err != nil {
		return nil, err
	}

	readDataLen := binary.BigEndian.Uint16(HeaderByte[:])
	if readDataLen+uint16(RpcHeaderLength) > uint16(reader.Length()) {
		return nil, nil
	}

	reader.Skip(RpcHeaderLength)
	bodyByte, err := reader.ReadBinary(int(readDataLen))
	if err != nil {
		return nil, err
	}

	return bodyByte, nil
}

func UnMarshalTStruct(ctx context.Context, iprot protocol.IProtocol, msg thrift.TStruct) (name string, seqid int32, err error) {
	if name, _, seqid, err = iprot.ReadMessageBegin(ctx); err != nil {
		return name, seqid, err
	}
	if err := msg.Read(ctx, iprot); err != nil {
		return name, seqid, err
	}
	if err := iprot.ReadMessageEnd(ctx); err != nil {
		return name, seqid, err
	}
	return name, seqid, err
}
