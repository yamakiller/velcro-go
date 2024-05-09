package remote

import (
	"errors"
	"fmt"
	"io"
	"sync"
)

// Mask bits.
const (
	BitReadable = 1 << iota
	BitWritable
)

var bytebufPool sync.Pool

func init() {
	bytebufPool.New = newDefaultByteBuffer
}

// NewWriterBuffer 创建一个指定大小的写入器缓冲区
// defaultByteBuffer 测试用用于验证其它缓冲区的正确性
func NewWriterBuffer(size int) ByteBuffer {
	return newWriterByteBuffer(size)
}

// NewReaderBuffer 创建一个固定的读取器
func NewReaderBuffer(buf []byte) ByteBuffer {
	return newReaderByteBuffer(buf)
}

// NewReaderWriterBuffer 创建一个指定大小的读写器
func NewReaderWriterBuffer(size int) ByteBuffer {
	return newReaderWriterByteBuffer(size)
}

type defaultByteBuffer struct {
	buff     []byte
	readIdx  int
	writeIdx int
	status   int
}

var _ ByteBuffer = &defaultByteBuffer{}

func newDefaultByteBuffer() interface{} {
	return &defaultByteBuffer{}
}

func newReaderByteBuffer(buf []byte) ByteBuffer {
	bytebuf := bytebufPool.Get().(*defaultByteBuffer)
	bytebuf.buff = buf
	bytebuf.readIdx = 0
	bytebuf.writeIdx = len(buf)
	bytebuf.status = BitReadable
	return bytebuf
}

func newWriterByteBuffer(estimatedLength int) ByteBuffer {
	if estimatedLength <= 0 {
		estimatedLength = 256 // default buffer size
	}
	bytebuf := bytebufPool.Get().(*defaultByteBuffer)
	bytebuf.buff = make([]byte, estimatedLength)
	bytebuf.status = BitWritable
	bytebuf.writeIdx = 0
	return bytebuf
}

func newReaderWriterByteBuffer(estimatedLength int) ByteBuffer {
	if estimatedLength <= 0 {
		estimatedLength = 256 // default buffer size
	}
	bytebuf := bytebufPool.Get().(*defaultByteBuffer)
	bytebuf.buff = make([]byte, estimatedLength)
	bytebuf.readIdx = 0
	bytebuf.writeIdx = 0
	bytebuf.status = BitReadable | BitWritable
	return bytebuf
}

func (b *defaultByteBuffer) Next(n int) (buf []byte, err error) {
	if b.status&BitReadable == 0 {
		return nil, errors.New("unreadable buffer, cannot support Next")
	}
	buf, err = b.Peek(n)
	b.readIdx += n
	return buf, err
}

// Peek ....
func (b *defaultByteBuffer) Peek(n int) (buf []byte, err error) {
	if b.status&BitReadable == 0 {
		return nil, errors.New("unreadable buffer, cannot support Peek")
	}
	if err = b.readableCheck(n); err != nil {
		return nil, err
	}
	return b.buff[b.readIdx : b.readIdx+n], nil
}

// Skip ...
func (b *defaultByteBuffer) Skip(n int) (err error) {
	if b.status&BitReadable == 0 {
		return errors.New("unreadable buffer, cannot support Skip")
	}
	if err = b.readableCheck(n); err != nil {
		return err
	}
	b.readIdx += n
	return nil
}

// ReadableLen ...
func (b *defaultByteBuffer) ReadableLen() (n int) {
	if b.status&BitReadable == 0 {
		return -1
	}
	return b.writeIdx - b.readIdx
}

// Read ...
func (b *defaultByteBuffer) Read(p []byte) (n int, err error) {
	if b.status&BitReadable == 0 {
		return -1, errors.New("unreadable buffer, cannot support Read")
	}
	pl := len(p)
	var buf []byte
	readable := b.ReadableLen()
	if readable == 0 {
		return 0, io.EOF
	}
	if pl <= readable {
		buf, err = b.Next(pl)
		if err != nil {
			return
		}
		n = pl
	} else {
		buf, err = b.Next(readable)
		if err != nil {
			return
		}
		n = readable
	}
	copy(p, buf)
	return
}

// ReadString ...
func (b *defaultByteBuffer) ReadString(n int) (s string, err error) {
	if b.status&BitReadable == 0 {
		return "", errors.New("unreadable buffer, cannot support ReadString")
	}
	var buf []byte
	if buf, err = b.Next(n); err != nil {
		return "", err
	}
	return string(buf), nil
}

// ReadBinary ...
func (b *defaultByteBuffer) ReadBinary(n int) (p []byte, err error) {
	if b.status&BitReadable == 0 {
		return p, errors.New("unreadable buffer, cannot support ReadBinary")
	}
	var buf []byte
	if buf, err = b.Next(n); err != nil {
		return p, err
	}
	p = make([]byte, n)
	copy(p, buf)
	return p, nil
}

// Malloc ...
func (b *defaultByteBuffer) Malloc(n int) (buf []byte, err error) {
	if b.status&BitWritable == 0 {
		return nil, errors.New("unwritable buffer, cannot support Malloc")
	}
	b.ensureWritable(n)
	currWIdx := b.writeIdx
	b.writeIdx += n
	return b.buff[currWIdx:b.writeIdx], nil
}

// MallocLen ...
func (b *defaultByteBuffer) MallocLen() (length int) {
	if b.status&BitWritable == 0 {
		return -1
	}
	return b.writeIdx
}

// WriteString ...
func (b *defaultByteBuffer) WriteString(s string) (n int, err error) {
	if b.status&BitWritable == 0 {
		return -1, errors.New("unwritable buffer, cannot support WriteString")
	}
	n = len(s)
	b.ensureWritable(n)
	copy(b.buff[b.writeIdx:b.writeIdx+n], s)
	b.writeIdx += n
	return
}

// Write ...
func (b *defaultByteBuffer) Write(p []byte) (n int, err error) {
	if b.status&BitWritable == 0 {
		return -1, errors.New("unwritable buffer, cannot support Write")
	}
	return b.WriteBinary(p)
}

// WriteBinary ...
func (b *defaultByteBuffer) WriteBinary(p []byte) (n int, err error) {
	if b.status&BitWritable == 0 {
		return -1, errors.New("unwritable buffer, cannot support WriteBinary")
	}
	n = len(p)
	b.ensureWritable(n)
	copy(b.buff[b.writeIdx:b.writeIdx+n], p)
	b.writeIdx += n
	return
}

// ReadLen ...
func (b *defaultByteBuffer) ReadLen() (n int) {
	return b.readIdx
}

// Flush ...
func (b *defaultByteBuffer) Flush() (err error) {
	if b.status&BitWritable == 0 {
		return errors.New("unwritable buffer, cannot support Flush")
	}
	return nil
}

// AppendBuffer ...
func (b *defaultByteBuffer) AppendBuffer(buf ByteBuffer) (err error) {
	subBuf := buf.(*defaultByteBuffer)
	n := subBuf.writeIdx
	b.ensureWritable(n)
	copy(b.buff[b.writeIdx:b.writeIdx+n], subBuf.buff)
	b.writeIdx += n
	buf.Release(nil)
	return
}

// Bytes ...
func (b *defaultByteBuffer) Bytes() (buf []byte, err error) {
	if b.status&BitWritable == 0 {
		return nil, errors.New("unwritable buffer, cannot support Bytes")
	}
	buf = make([]byte, b.writeIdx)
	copy(buf, b.buff[:b.writeIdx])
	return buf, nil
}

// NewBuffer 船舰一个256字节的缓冲区
func (b *defaultByteBuffer) NewBuffer() ByteBuffer {
	return NewWriterBuffer(256)
}

// Release 释放缓冲区
func (b *defaultByteBuffer) Release(e error) (err error) {
	b.zero()
	bytebufPool.Put(b)
	return
}

func (b *defaultByteBuffer) zero() {
	b.status = 0
	b.buff = nil
	b.readIdx = 0
	b.writeIdx = 0
}

func (b *defaultByteBuffer) readableCheck(n int) error {
	readableLen := b.ReadableLen()
	if b.readIdx >= b.writeIdx && n > 0 {
		return io.EOF
	}
	if readableLen < n {
		return fmt.Errorf("default byteBuffer not enough, read[%d] but remain[%d]", n, readableLen)
	}
	return nil
}

func (b *defaultByteBuffer) writableLen() (n int) {
	if b.status&BitWritable == 0 {
		return -1
	}
	return cap(b.buff) - b.writeIdx
}

func (b *defaultByteBuffer) ensureWritable(minWritableBytes int) {
	if b.writableLen() >= minWritableBytes {
		return
	}
	minNewCapacity := b.writeIdx + minWritableBytes
	newCapacity := cap(b.buff)
	for newCapacity < minNewCapacity {
		newCapacity <<= 1
	}

	buf := make([]byte, newCapacity)
	copy(buf, b.buff)
	b.buff = buf
}
