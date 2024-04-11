package protocol

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/utils/circbuf"
)

func NewBinaryProtocol()*BinaryProtocol{
	return &BinaryProtocol{
		trans: NewTransLink(),
	}
}

type BinaryProtocol struct {
	trans  *TransLink
	cfg    *thrift.TConfiguration
	buffer [64]byte
	transport Transport
}

func (p BinaryProtocol) GetBytes()[]byte{
	b :=p.trans.Bytes()
	res := make([]byte,len(b))
	copy(res,b)
	return res
}

func (p *BinaryProtocol) Write(b[]byte) (int, error){
	return p.trans.Write(b)
}
func (p *BinaryProtocol) Reader() circbuf.Reader{
	return p.trans
}
func (p *BinaryProtocol)Release(){
	p.trans.Close()
	p.trans = NewTransLink()
}
func (p *BinaryProtocol) Close(){
	p.trans.Close()
}
func (p *BinaryProtocol) Flush(ctx context.Context) (err error){
	return nil
}
func (p *BinaryProtocol) SetTConfiguration(conf *thrift.TConfiguration){

}
func (p *BinaryProtocol) Skip(ctx context.Context, fieldType thrift.TType) (err error){
	return nil
}

func (p *BinaryProtocol) Transport() thrift.TTransport{
	return &p.transport
}
/**
 * Writing Methods
 */

func (p *BinaryProtocol) WriteMessageBegin(ctx context.Context, name string, typeId thrift.TMessageType, seqId int32) error {
	if p.cfg.GetTBinaryStrictWrite() {
		version := uint32(thrift.VERSION_1) | uint32(typeId)
		e := p.WriteI32(ctx, int32(version))
		if e != nil {
			return e
		}
		e = p.WriteString(ctx, name)
		if e != nil {
			return e
		}
		e = p.WriteI32(ctx, seqId)
		return e
	} else {
		e := p.WriteString(ctx, name)
		if e != nil {
			return e
		}
		e = p.WriteByte(ctx, int8(typeId))
		if e != nil {
			return e
		}
		e = p.WriteI32(ctx, seqId)
		return e
	}
}

func (p *BinaryProtocol) WriteMessageEnd(ctx context.Context) error {
	return p.trans.Flush()
}

func (p *BinaryProtocol) WriteStructBegin(ctx context.Context, name string) error {
	return nil
}

func (p *BinaryProtocol) WriteStructEnd(ctx context.Context) error {
	return nil
}

func (p *BinaryProtocol) WriteFieldBegin(ctx context.Context, name string, typeId thrift.TType, id int16) error {
	e := p.WriteByte(ctx, int8(typeId))
	if e != nil {
		return e
	}
	e = p.WriteI16(ctx, id)
	return e
}

func (p *BinaryProtocol) WriteFieldEnd(ctx context.Context) error {
	return nil
}

func (p *BinaryProtocol) WriteFieldStop(ctx context.Context) error {
	e := p.WriteByte(ctx, thrift.STOP)
	return e
}

func (p *BinaryProtocol) WriteMapBegin(ctx context.Context, keyType thrift.TType, valueType thrift.TType, size int) error {
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

func (p *BinaryProtocol) WriteMapEnd(ctx context.Context) error {
	return nil
}

func (p *BinaryProtocol) WriteListBegin(ctx context.Context, elemType thrift.TType, size int) error {
	e := p.WriteByte(ctx, int8(elemType))
	if e != nil {
		return e
	}
	e = p.WriteI32(ctx, int32(size))
	return e
}

func (p *BinaryProtocol) WriteListEnd(ctx context.Context) error {
	return nil
}

func (p *BinaryProtocol) WriteSetBegin(ctx context.Context, elemType thrift.TType, size int) error {
	e := p.WriteByte(ctx, int8(elemType))
	if e != nil {
		return e
	}
	e = p.WriteI32(ctx, int32(size))
	return e
}

func (p *BinaryProtocol) WriteSetEnd(ctx context.Context) error {
	return nil
}

func (p *BinaryProtocol) WriteBool(ctx context.Context, value bool) error {
	if value {
		return p.WriteByte(ctx, 1)
	}
	return p.WriteByte(ctx, 0)
}

func (p *BinaryProtocol) WriteByte(ctx context.Context, value int8) error {
	e := p.trans.WriteByte(byte(value))
	return thrift.NewTProtocolException(e)
}

func (p *BinaryProtocol) WriteI16(ctx context.Context, value int16) error {
	v := p.buffer[0:2]
	binary.BigEndian.PutUint16(v, uint16(value))
	_, e := p.trans.Write(v)
	return thrift.NewTProtocolException(e)
}

func (p *BinaryProtocol) WriteI32(ctx context.Context, value int32) error {
	v := p.buffer[0:4]
	binary.BigEndian.PutUint32(v, uint32(value))
	_, e := p.trans.Write(v)
	return thrift.NewTProtocolException(e)
}

func (p *BinaryProtocol) WriteI64(ctx context.Context, value int64) error {
	v := p.buffer[0:8]
	binary.BigEndian.PutUint64(v, uint64(value))
	_, err := p.trans.Write(v)
	return thrift.NewTProtocolException(err)
}

func (p *BinaryProtocol) WriteDouble(ctx context.Context, value float64) error {
	return p.WriteI64(ctx, int64(math.Float64bits(value)))
}

func (p *BinaryProtocol) WriteString(ctx context.Context, value string) error {
	e := p.WriteI32(ctx, int32(len(value)))
	if e != nil {
		return e
	}
	_, err := p.trans.WriteString(value)
	return thrift.NewTProtocolException(err)
}

func (p *BinaryProtocol) WriteBinary(ctx context.Context, value []byte) error {
	e := p.WriteI32(ctx, int32(len(value)))
	if e != nil {
		return e
	}
	_, err := p.trans.Write(value)
	return thrift.NewTProtocolException(err)
}

func (p *BinaryProtocol) WriteUUID(ctx context.Context, value thrift.Tuuid) error {
	_, err := p.trans.Write(value[:])
	return thrift.NewTProtocolException(err)
}

/**
 * Reading methods
 */

 func (p *BinaryProtocol) ReadMessageBegin(ctx context.Context) (name string, typeId thrift.TMessageType, seqId int32, err error) {
	size, e := p.ReadI32(ctx)
	if e != nil {
		return "", typeId, 0, thrift.NewTProtocolException(e)
	}
	if size < 0 {
		typeId = thrift.TMessageType(size & 0x0ff)
		version := int64(int64(size) & thrift.VERSION_MASK)
		if version != thrift.VERSION_1 {
			return name, typeId, seqId, thrift.NewTProtocolExceptionWithType(thrift.BAD_VERSION, fmt.Errorf("Bad version in ReadMessageBegin"))
		}
		name, e = p.ReadString(ctx)
		if e != nil {
			return name, typeId, seqId, thrift.NewTProtocolException(e)
		}
		seqId, e = p.ReadI32(ctx)
		if e != nil {
			return name, typeId, seqId, thrift.NewTProtocolException(e)
		}
		return name, typeId, seqId, nil
	}
	if p.cfg.GetTBinaryStrictRead() {
		return name, typeId, seqId, thrift.NewTProtocolExceptionWithType(thrift.BAD_VERSION, fmt.Errorf("Missing version in ReadMessageBegin"))
	}
	name, e2 := p.readStringBody(size)
	if e2 != nil {
		return name, typeId, seqId, e2
	}
	b, e3 := p.ReadByte(ctx)
	if e3 != nil {
		return name, typeId, seqId, e3
	}
	typeId = thrift.TMessageType(b)
	seqId, e4 := p.ReadI32(ctx)
	if e4 != nil {
		return name, typeId, seqId, e4
	}
	return name, typeId, seqId, nil
}

func (p *BinaryProtocol) ReadMessageEnd(ctx context.Context) error {
	return nil
}

func (p *BinaryProtocol) ReadStructBegin(ctx context.Context) (name string, err error) {
	return
}

func (p *BinaryProtocol) ReadStructEnd(ctx context.Context) error {
	return nil
}

func (p *BinaryProtocol) ReadFieldBegin(ctx context.Context) (name string, typeId thrift.TType, seqId int16, err error) {
	t, err := p.ReadByte(ctx)
	typeId = thrift.TType(t)
	if err != nil {
		return name, typeId, seqId, err
	}
	if t != thrift.STOP {
		seqId, err = p.ReadI16(ctx)
	}
	return name, typeId, seqId, err
}

func (p *BinaryProtocol) ReadFieldEnd(ctx context.Context) error {
	return nil
}

func (p *BinaryProtocol) ReadMapBegin(ctx context.Context) (kType, vType thrift.TType, size int, err error) {
	k, e := p.ReadByte(ctx)
	if e != nil {
		err = thrift.NewTProtocolException(e)
		return
	}
	kType = thrift.TType(k)
	v, e := p.ReadByte(ctx)
	if e != nil {
		err = thrift.NewTProtocolException(e)
		return
	}
	vType = thrift.TType(v)
	size32, e := p.ReadI32(ctx)
	if e != nil {
		err = thrift.NewTProtocolException(e)
		return
	}
	err = checkSizeForProtocol(size32, p.cfg)
	if err != nil {
		return
	}
	size = int(size32)
	return kType, vType, size, nil
}

func (p *BinaryProtocol) ReadMapEnd(ctx context.Context) error {
	return nil
}

func (p *BinaryProtocol) ReadListBegin(ctx context.Context) (elemType thrift.TType, size int, err error) {
	b, e := p.ReadByte(ctx)
	if e != nil {
		err = thrift.NewTProtocolException(e)
		return
	}
	elemType = thrift.TType(b)
	size32, e := p.ReadI32(ctx)
	if e != nil {
		err = thrift.NewTProtocolException(e)
		return
	}
	err = checkSizeForProtocol(size32, p.cfg)
	if err != nil {
		return
	}
	size = int(size32)

	return
}

func (p *BinaryProtocol) ReadListEnd(ctx context.Context) error {
	return nil
}

func (p *BinaryProtocol) ReadSetBegin(ctx context.Context) (elemType thrift.TType, size int, err error) {
	b, e := p.ReadByte(ctx)
	if e != nil {
		err = thrift.NewTProtocolException(e)
		return
	}
	elemType = thrift.TType(b)
	size32, e := p.ReadI32(ctx)
	if e != nil {
		err = thrift.NewTProtocolException(e)
		return
	}
	err = checkSizeForProtocol(size32, p.cfg)
	if err != nil {
		return
	}
	size = int(size32)
	return elemType, size, nil
}

func (p *BinaryProtocol) ReadSetEnd(ctx context.Context) error {
	return nil
}

func (p *BinaryProtocol) ReadBool(ctx context.Context) (bool, error) {
	b, e := p.ReadByte(ctx)
	v := true
	if b != 1 {
		v = false
	}
	return v, e
}

func (p *BinaryProtocol) ReadByte(ctx context.Context) (int8, error) {
	v, err := p.trans.ReadByte()
	return int8(v), err
}

func (p *BinaryProtocol) ReadI16(ctx context.Context) (value int16, err error) {
	buf := p.buffer[0:2]
	err = p.readAll(ctx, buf)
	value = int16(binary.BigEndian.Uint16(buf))
	return value, err
}

func (p *BinaryProtocol) ReadI32(ctx context.Context) (value int32, err error) {
	buf := p.buffer[0:4]
	err = p.readAll(ctx, buf)
	value = int32(binary.BigEndian.Uint32(buf))
	return value, err
}

func (p *BinaryProtocol) ReadI64(ctx context.Context) (value int64, err error) {
	buf := p.buffer[0:8]
	err = p.readAll(ctx, buf)
	value = int64(binary.BigEndian.Uint64(buf))
	return value, err
}

func (p *BinaryProtocol) ReadDouble(ctx context.Context) (value float64, err error) {
	buf := p.buffer[0:8]
	err = p.readAll(ctx, buf)
	value = math.Float64frombits(binary.BigEndian.Uint64(buf))
	return value, err
}

func (p *BinaryProtocol) ReadString(ctx context.Context) (value string, err error) {
	size, e := p.ReadI32(ctx)
	if e != nil {
		return "", e
	}
	err = checkSizeForProtocol(size, p.cfg)
	if err != nil {
		return
	}
	if size == 0 {
		return "", nil
	}
	return p.readStringBody(size)
}

func (p *BinaryProtocol) ReadBinary(ctx context.Context) ([]byte, error) {
	size, e := p.ReadI32(ctx)
	if e != nil {
		return nil, e
	}
	if err := checkSizeForProtocol(size, p.cfg); err != nil {
		return nil, err
	}

	buf, err := p.trans.ReadBinary(int(size))
	return buf, thrift.NewTProtocolException(err)
}

func (p *BinaryProtocol) ReadUUID(ctx context.Context) (value thrift.Tuuid, err error) {
	buf := p.buffer[0:16]
	err = p.readAll(ctx, buf)
	if err == nil {
		copy(value[:], buf)
	}
	return value, err
}



func (p *BinaryProtocol) readAll(ctx context.Context, buf []byte) (err error) {
	b ,err := p.trans.ReadBinary(len(buf))
	copy(buf,b)
	return nil
}

func (p *BinaryProtocol) readStringBody(size int32) (value string, err error) {
	buf, err := p.trans.ReadBinary(int(size))
	return string(buf),thrift. NewTProtocolException(err)
}

// type timeoutable interface {
// 	Timeout() bool
// }


// func isTimeoutError(err error) bool {
// 	var t timeoutable
// 	if errors.As(err, &t) {
// 		return t.Timeout()
// 	}
// 	return false
// }

func checkSizeForProtocol(size int32, cfg *thrift.TConfiguration) error {
	if size < 0 {
		return thrift.NewTProtocolExceptionWithType(
			thrift.NEGATIVE_SIZE,
			fmt.Errorf("negative size: %d", size),
		)
	}
	if size > cfg.GetMaxMessageSize() {
		return thrift.NewTProtocolExceptionWithType(
			thrift.SIZE_LIMIT,
			fmt.Errorf("size exceeded max allowed: %d", size),
		)
	}
	return nil
}