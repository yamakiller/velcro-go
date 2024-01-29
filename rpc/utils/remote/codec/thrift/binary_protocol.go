package thrift

import (
	"context"
	"encoding/binary"
	"math"
	"sync"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/rpc/utils/remote"
	"github.com/yamakiller/velcro-go/rpc/utils/remote/codec/perrors"
)

// must be strict read & strict write
var (
	bpPool sync.Pool
	_      thrift.TProtocol = (*BinaryProtocol)(nil)
)

func init() {
	bpPool.New = newBP
}

func newBP() interface{} {
	return &BinaryProtocol{}
}

// NewBinaryProtocol ...
func NewBinaryProtocol(t remote.ByteBuffer) *BinaryProtocol {
	bp := bpPool.Get().(*BinaryProtocol)
	bp.trans = t
	return bp
}

// BinaryProtocol ...
type BinaryProtocol struct {
	trans remote.ByteBuffer
}

// Recycle ...
func (p *BinaryProtocol) Recycle() {
	p.trans = nil
	bpPool.Put(p)
}

/**
 * Writing Methods
 */

// WriteMessageBegin ...
func (p *BinaryProtocol) WriteMessageBegin(ctx context.Context, name string, typeID thrift.TMessageType, seqID int32) error {
	version := uint32(thrift.VERSION_1) | uint32(typeID)
	e := p.WriteI32(ctx, int32(version))
	if e != nil {
		return e
	}
	e = p.WriteString(ctx, name)
	if e != nil {
		return e
	}
	e = p.WriteI32(ctx, seqID)
	return e
}

// WriteMessageEnd ...
func (p *BinaryProtocol) WriteMessageEnd(ctx context.Context) error {
	return nil
}

// WriteStructBegin ...
func (p *BinaryProtocol) WriteStructBegin(ctx context.Context, name string) error {
	return nil
}

// WriteStructEnd ...
func (p *BinaryProtocol) WriteStructEnd(ctx context.Context) error {
	return nil
}

// WriteFieldBegin ...
func (p *BinaryProtocol) WriteFieldBegin(ctx context.Context, name string, typeID thrift.TType, id int16) error {
	e := p.WriteByte(ctx, int8(typeID))
	if e != nil {
		return e
	}
	e = p.WriteI16(ctx, id)
	return e
}

// WriteFieldEnd ...
func (p *BinaryProtocol) WriteFieldEnd(ctx context.Context) error {
	return nil
}

// WriteFieldStop ...
func (p *BinaryProtocol) WriteFieldStop(ctx context.Context) error {
	e := p.WriteByte(ctx, thrift.STOP)
	return e
}

// WriteMapBegin ...
func (p *BinaryProtocol) WriteMapBegin(ctx context.Context, keyType, valueType thrift.TType, size int) error {
	e := p.WriteByte(ctx, int8(keyType))
	if e != nil {
		return e
	}
	e = p.WriteByte(ctx, int8(valueType))
	if e != nil {
		return e
	}
	e = p.WriteI32(ctx, int32(size))
	return e
}

// WriteMapEnd ...
func (p *BinaryProtocol) WriteMapEnd(ctx context.Context) error {
	return nil
}

// WriteListBegin ...
func (p *BinaryProtocol) WriteListBegin(ctx context.Context, elemType thrift.TType, size int) error {
	e := p.WriteByte(ctx, int8(elemType))
	if e != nil {
		return e
	}
	e = p.WriteI32(ctx, int32(size))
	return e
}

// WriteListEnd ...
func (p *BinaryProtocol) WriteListEnd(ctx context.Context) error {
	return nil
}

// WriteSetBegin ...
func (p *BinaryProtocol) WriteSetBegin(ctx context.Context, elemType thrift.TType, size int) error {
	e := p.WriteByte(ctx, int8(elemType))
	if e != nil {
		return e
	}
	e = p.WriteI32(ctx, int32(size))
	return e
}

// WriteSetEnd ...
func (p *BinaryProtocol) WriteSetEnd(ctx context.Context) error {
	return nil
}

// WriteBool ...
func (p *BinaryProtocol) WriteBool(ctx context.Context, value bool) error {
	if value {
		return p.WriteByte(ctx, 1)
	}
	return p.WriteByte(ctx, 0)
}

// WriteByte ...
func (p *BinaryProtocol) WriteByte(ctx context.Context, value int8) error {
	v, err := p.malloc(1)
	if err != nil {
		return err
	}
	v[0] = byte(value)
	return err
}

// WriteI16 ...
func (p *BinaryProtocol) WriteI16(ctx context.Context, value int16) error {
	v, err := p.malloc(2)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint16(v, uint16(value))
	return err
}

// WriteI32 ...
func (p *BinaryProtocol) WriteI32(ctx context.Context, value int32) error {
	v, err := p.malloc(4)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint32(v, uint32(value))
	return err
}

// WriteI64 ...
func (p *BinaryProtocol) WriteI64(ctx context.Context, value int64) error {
	v, err := p.malloc(8)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint64(v, uint64(value))
	return err
}

// WriteDouble ...
func (p *BinaryProtocol) WriteDouble(ctx context.Context, value float64) error {
	return p.WriteI64(ctx, int64(math.Float64bits(value)))
}

// WriteString ...
func (p *BinaryProtocol) WriteString(ctx context.Context, value string) error {
	len := len(value)
	e := p.WriteI32(ctx, int32(len))
	if e != nil {
		return e
	}
	_, e = p.trans.WriteString(value)
	return e
}

// WriteBinary ...
func (p *BinaryProtocol) WriteBinary(ctx context.Context, value []byte) error {
	e := p.WriteI32(ctx, int32(len(value)))
	if e != nil {
		return e
	}
	_, e = p.trans.WriteBinary(value)
	return e
}

func (p *BinaryProtocol) WriteUUID(ctx context.Context, value thrift.Tuuid) error {
	_, e := p.trans.Write(value[:])
	return e
}

// malloc ...
func (p *BinaryProtocol) malloc(size int) ([]byte, error) {
	buf, err := p.trans.Malloc(size)
	if err != nil {
		return buf, perrors.NewProtocolError(err)
	}
	return buf, nil
}

/**
 * Reading methods
 */

// ReadMessageBegin ...
func (p *BinaryProtocol) ReadMessageBegin(ctx context.Context) (name string, typeID thrift.TMessageType, seqID int32, err error) {
	size, e := p.ReadI32(ctx)
	if e != nil {
		return "", typeID, 0, perrors.NewProtocolError(e)
	}
	if size > 0 {
		return name, typeID, seqID, perrors.NewProtocolErrorWithType(perrors.BadVersion, "Missing version in ReadMessageBegin")
	}
	typeID = thrift.TMessageType(size & 0x0ff)
	version := int64(int64(size) & thrift.VERSION_MASK)
	if version != thrift.VERSION_1 {
		return name, typeID, seqID, perrors.NewProtocolErrorWithType(perrors.BadVersion, "Bad version in ReadMessageBegin")
	}
	name, e = p.ReadString(ctx)
	if e != nil {
		return name, typeID, seqID, perrors.NewProtocolError(e)
	}
	seqID, e = p.ReadI32(ctx)
	if e != nil {
		return name, typeID, seqID, perrors.NewProtocolError(e)
	}
	return name, typeID, seqID, nil
}

// ReadMessageEnd ...
func (p *BinaryProtocol) ReadMessageEnd(ctx context.Context) error {
	return nil
}

// ReadStructBegin ...
func (p *BinaryProtocol) ReadStructBegin(ctx context.Context) (name string, err error) {
	return
}

// ReadStructEnd ...
func (p *BinaryProtocol) ReadStructEnd(ctx context.Context) error {
	return nil
}

// ReadFieldBegin ...
func (p *BinaryProtocol) ReadFieldBegin(ctx context.Context) (name string, typeID thrift.TType, id int16, err error) {
	t, err := p.ReadByte(ctx)
	typeID = thrift.TType(t)
	if err != nil {
		return name, typeID, id, err
	}
	if t != thrift.STOP {
		id, err = p.ReadI16(ctx)
	}
	return name, typeID, id, err
}

// ReadFieldEnd ...
func (p *BinaryProtocol) ReadFieldEnd(ctx context.Context) error {
	return nil
}

// ReadMapBegin ...
func (p *BinaryProtocol) ReadMapBegin(ctx context.Context) (kType, vType thrift.TType, size int, err error) {
	k, e := p.ReadByte(ctx)
	if e != nil {
		err = perrors.NewProtocolError(e)
		return
	}
	kType = thrift.TType(k)
	v, e := p.ReadByte(ctx)
	if e != nil {
		err = perrors.NewProtocolError(e)
		return
	}
	vType = thrift.TType(v)
	size32, e := p.ReadI32(ctx)
	if e != nil {
		err = perrors.NewProtocolError(e)
		return
	}
	if size32 < 0 {
		err = perrors.InvalidDataLength
		return
	}
	size = int(size32)
	return kType, vType, size, nil
}

// ReadMapEnd ...
func (p *BinaryProtocol) ReadMapEnd(ctx context.Context) error {
	return nil
}

// ReadListBegin ...
func (p *BinaryProtocol) ReadListBegin(ctx context.Context) (elemType thrift.TType, size int, err error) {
	b, e := p.ReadByte(ctx)
	if e != nil {
		err = perrors.NewProtocolError(e)
		return
	}
	elemType = thrift.TType(b)
	size32, e := p.ReadI32(ctx)
	if e != nil {
		err = perrors.NewProtocolError(e)
		return
	}
	if size32 < 0 {
		err = perrors.InvalidDataLength
		return
	}
	size = int(size32)

	return
}

// ReadListEnd ...
func (p *BinaryProtocol) ReadListEnd(ctx context.Context) error {
	return nil
}

// ReadSetBegin ...
func (p *BinaryProtocol) ReadSetBegin(ctx context.Context) (elemType thrift.TType, size int, err error) {
	b, e := p.ReadByte(ctx)
	if e != nil {
		err = perrors.NewProtocolError(e)
		return
	}
	elemType = thrift.TType(b)
	size32, e := p.ReadI32(ctx)
	if e != nil {
		err = perrors.NewProtocolError(e)
		return
	}
	if size32 < 0 {
		err = perrors.InvalidDataLength
		return
	}
	size = int(size32)
	return elemType, size, nil
}

// ReadSetEnd ...
func (p *BinaryProtocol) ReadSetEnd(ctx context.Context) error {
	return nil
}

// ReadBool ...
func (p *BinaryProtocol) ReadBool(ctx context.Context) (bool, error) {
	b, e := p.ReadByte(ctx)
	v := true
	if b != 1 {
		v = false
	}
	return v, e
}

// ReadByte ...
func (p *BinaryProtocol) ReadByte(ctx context.Context) (value int8, err error) {
	buf, err := p.next(1)
	if err != nil {
		return value, err
	}
	return int8(buf[0]), err
}

// ReadI16 ...
func (p *BinaryProtocol) ReadI16(ctx context.Context) (value int16, err error) {
	buf, err := p.next(2)
	if err != nil {
		return value, err
	}
	value = int16(binary.BigEndian.Uint16(buf))
	return value, err
}

// ReadI32 ...
func (p *BinaryProtocol) ReadI32(ctx context.Context) (value int32, err error) {
	buf, err := p.next(4)
	if err != nil {
		return value, err
	}
	value = int32(binary.BigEndian.Uint32(buf))
	return value, err
}

// ReadI64 ...
func (p *BinaryProtocol) ReadI64(ctx context.Context) (value int64, err error) {
	buf, err := p.next(8)
	if err != nil {
		return value, err
	}
	value = int64(binary.BigEndian.Uint64(buf))
	return value, err
}

// ReadDouble ...
func (p *BinaryProtocol) ReadDouble(ctx context.Context) (value float64, err error) {
	buf, err := p.next(8)
	if err != nil {
		return value, err
	}
	value = math.Float64frombits(binary.BigEndian.Uint64(buf))
	return value, err
}

// ReadString ...
func (p *BinaryProtocol) ReadString(ctx context.Context) (value string, err error) {
	size, e := p.ReadI32(ctx)
	if e != nil {
		return "", e
	}
	if size < 0 {
		err = perrors.InvalidDataLength
		return
	}
	value, err = p.trans.ReadString(int(size))
	if err != nil {
		return value, perrors.NewProtocolError(err)
	}
	return value, nil
}

// ReadBinary ...
func (p *BinaryProtocol) ReadBinary(ctx context.Context) ([]byte, error) {
	size, e := p.ReadI32(ctx)
	if e != nil {
		return nil, e
	}
	if size < 0 {
		return nil, perrors.InvalidDataLength
	}
	return p.trans.ReadBinary(int(size))
}

func (p *BinaryProtocol) ReadUUID(ctx context.Context) (value thrift.Tuuid, err error) {
	var buf [16]byte
	_ ,e := p.trans.Read(buf[:])
	if e != nil {
		return value, e
	}
	copy(value[:], buf[:])
	return value, err
}

func (p *BinaryProtocol) Flush(ctx context.Context) (err error) {
	err = p.trans.Flush()
	if err != nil {
		return perrors.NewProtocolError(err)
	}
	return nil
}

// Skip ...
func (p *BinaryProtocol) Skip(ctx context.Context, fieldType thrift.TType) (err error) {
	return thrift.SkipDefaultDepth(ctx, p, fieldType)
}

// Transport ...
func (p *BinaryProtocol) Transport() thrift.TTransport {
	// not support
	return nil
}

// ByteBuffer ...
func (p *BinaryProtocol) ByteBuffer() remote.ByteBuffer {
	return p.trans
}

// next ...
func (p *BinaryProtocol) next(size int) ([]byte, error) {
	buf, err := p.trans.Next(size)
	if err != nil {
		return buf, perrors.NewProtocolError(err)
	}
	return buf, nil
}
