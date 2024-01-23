//go:build amd64 && !windows && go1.16 && !go1.22 && !disablefrugal
// +build amd64,!windows,go1.16,!go1.22,!disablefrugal

package thrift

import (
	"fmt"
	"reflect"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/rpc/utils/protocol/bthrift"
	"github.com/yamakiller/velcro-go/rpc/utils/remote/codec/perrors"
)

const (
	// 0b0001 and 0b0010 are used for FastWrite and FastRead, so Frugal starts from 0b0100
	FrugalWrite CodecType = 0b0100
	FrugalRead  CodecType = 0b1000

	FrugalReadWrite = FrugalWrite | FrugalRead
)

// hyperMarshalEnabled indicates that if there are high priority message codec for current platform.
func (c thriftCodec) hyperMarshalEnabled() bool {
	return c.CodecType&FrugalWrite != 0
}

// hyperMarshalAvailable indicates that if high priority message codec is available.
func hyperMarshalAvailable(data interface{}) bool {
	dt := reflect.TypeOf(data).Elem()
	if dt.NumField() > 0 && dt.Field(0).Tag.Get("frugal") == "" {
		return false
	}
	return true
}

// hyperMessageUnmarshalEnabled indicates that if there are high priority message codec for current platform.
func (c thriftCodec) hyperMessageUnmarshalEnabled() bool {
	return c.CodecType&FrugalRead != 0
}

// hyperMessageUnmarshalAvailable indicates that if high priority message codec is available.
func hyperMessageUnmarshalAvailable(data interface{}, payloadLen int) bool {
	if payloadLen == 0 {
		return false
	}
	dt := reflect.TypeOf(data).Elem()
	if dt.NumField() > 0 && dt.Field(0).Tag.Get("frugal") == "" {
		return false
	}
	return true
}

func (c thriftCodec) hyperMarshal(out remote.ByteBuffer, methodName string, msgType remote.MessageType,
	seqID int32, data interface{},
) error {
	// calculate and malloc message buffer
	msgBeginLen := bthrift.Binary.MessageBeginLength(methodName, thrift.TMessageType(msgType), seqID)
	msgEndLen := bthrift.Binary.MessageEndLength()
	objectLen := frugal.EncodedSize(data)
	buf, err := out.Malloc(msgBeginLen + objectLen + msgEndLen)
	if err != nil {
		return perrors.NewProtocolErrorWithMsg(fmt.Sprintf("thrift marshal, Malloc failed: %s", err.Error()))
	}

	// encode message
	offset := bthrift.Binary.WriteMessageBegin(buf, methodName, thrift.TMessageType(msgType), seqID)
	var writeLen int
	writeLen, err = frugal.EncodeObject(buf[offset:], nil, data)
	if err != nil {
		return perrors.NewProtocolErrorWithMsg(fmt.Sprintf("thrift marshal, Encode failed: %s", err.Error()))
	}
	offset += writeLen
	bthrift.Binary.WriteMessageEnd(buf[offset:])
	return nil
}

func (c thriftCodec) hyperMarshalBody(data interface{}) (buf []byte, err error) {
	objectLen := frugal.EncodedSize(data)
	buf = mcache.Malloc(objectLen)
	_, err = frugal.EncodeObject(buf, nil, data)
	return buf, err
}

func (c thriftCodec) hyperMessageUnmarshal(buf []byte, data interface{}) error {
	_, err := frugal.DecodeObject(buf, data)
	if err != nil {
		return remote.NewTransError(remote.ProtocolError, err)
	}
	return nil
}
