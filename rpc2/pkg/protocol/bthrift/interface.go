// Package bthrift is byted thrift
package bthrift

import (
	"context"

	"github.com/apache/thrift/lib/go/thrift"
)

// BinaryWriter .
type BinaryWriter interface {
	WriteDirect(b []byte, remainCap int) error
}

// BTProtocol .
type BTProtocol interface {
	WriteMessageBegin(ctx context.Context, buf []byte, name string, typeID thrift.TMessageType, seqid int32) int
	WriteMessageEnd(ctx context.Context, buf []byte) int
	WriteStructBegin(ctx context.Context, buf []byte, name string) int
	WriteStructEnd(ctx context.Context, buf []byte) int
	WriteFieldBegin(ctx context.Context, buf []byte, name string, typeID thrift.TType, id int16) int
	WriteFieldEnd(ctx context.Context, buf []byte) int
	WriteFieldStop(ctx context.Context, buf []byte) int
	WriteMapBegin(ctx context.Context, buf []byte, keyType, valueType thrift.TType, size int) int
	WriteMapEnd(ctx context.Context, buf []byte) int
	WriteListBegin(ctx context.Context, buf []byte, elemType thrift.TType, size int) int
	WriteListEnd(ctx context.Context, buf []byte) int
	WriteSetBegin(ctx context.Context, buf []byte, elemType thrift.TType, size int) int
	WriteSetEnd(ctx context.Context, buf []byte) int
	WriteBool(ctx context.Context, buf []byte, value bool) int
	WriteByte(ctx context.Context, buf []byte, value int8) int
	WriteI16(ctx context.Context, buf []byte, value int16) int
	WriteI32(ctx context.Context, buf []byte, value int32) int
	WriteI64(ctx context.Context, buf []byte, value int64) int
	WriteDouble(ctx context.Context, buf []byte, value float64) int
	WriteString(ctx context.Context, buf []byte, value string) int
	WriteBinary(ctx context.Context, buf, value []byte) int
	WriteStringNocopy(ctx context.Context, buf []byte, binaryWriter BinaryWriter, value string) int
	WriteBinaryNocopy(ctx context.Context, buf []byte, binaryWriter BinaryWriter, value []byte) int
	MessageBeginLength(name string, typeID thrift.TMessageType, seqid int32) int
	MessageEndLength() int
	StructBeginLength(name string) int
	StructEndLength() int
	FieldBeginLength(ctx context.Context, name string, typeID thrift.TType, id int16) int
	FieldEndLength(ctx context.Context) int
	FieldStopLength(ctx context.Context) int
	MapBeginLength(ctx context.Context, keyType, valueType thrift.TType, size int) int
	MapEndLength(ctx context.Context) int
	ListBeginLength(ctx context.Context, elemType thrift.TType, size int) int
	ListEndLength(ctx context.Context) int
	SetBeginLength(ctx context.Context, elemType thrift.TType, size int) int
	SetEndLength(ctx context.Context) int
	BoolLength(ctx context.Context, value bool) int
	ByteLength(ctx context.Context, value int8) int
	I16Length(ctx context.Context, value int16) int
	I32Length(ctx context.Context, value int32) int
	I64Length(ctx context.Context, value int64) int
	DoubleLength(value float64) int
	StringLength(value string) int
	BinaryLength(value []byte) int
	StringLengthNocopy(ctx context.Context, value string) int
	BinaryLengthNocopy(ctx context.Context, value []byte) int
	ReadMessageBegin(ctx context.Context, buf []byte) (name string, typeID thrift.TMessageType, seqid int32, length int, err error)
	ReadMessageEnd(ctx context.Context, buf []byte) (int, error)
	ReadStructBegin(ctx context.Context, buf []byte) (name string, length int, err error)
	ReadStructEnd(ctx context.Context, buf []byte) (int, error)
	ReadFieldBegin(ctx context.Context, buf []byte) (name string, typeID thrift.TType, id int16, length int, err error)
	ReadFieldEnd(ctx context.Context, buf []byte) (int, error)
	ReadMapBegin(ctx context.Context, buf []byte) (keyType, valueType thrift.TType, size, length int, err error)
	ReadMapEnd(ctx context.Context, buf []byte) (int, error)
	ReadListBegin(ctx context.Context, buf []byte) (elemType thrift.TType, size, length int, err error)
	ReadListEnd(ctx context.Context, buf []byte) (int, error)
	ReadSetBegin(ctx context.Context, buf []byte) (elemType thrift.TType, size, length int, err error)
	ReadSetEnd(ctx context.Context, buf []byte) (int, error)
	ReadBool(ctx context.Context, buf []byte) (value bool, length int, err error)
	ReadByte(ctx context.Context, buf []byte) (value int8, length int, err error)
	ReadI16(ctx context.Context, buf []byte) (value int16, length int, err error)
	ReadI32(ctx context.Context, buf []byte) (value int32, length int, err error)
	ReadI64(ctx context.Context, buf []byte) (value int64, length int, err error)
	ReadDouble(ctx context.Context, buf []byte) (value float64, length int, err error)
	ReadString(ctx context.Context, buf []byte) (value string, length int, err error)
	ReadBinary(ctx context.Context, buf []byte) (value []byte, length int, err error)
	Skip(ctx context.Context, buf []byte, fieldType thrift.TType) (length int, err error)
}
