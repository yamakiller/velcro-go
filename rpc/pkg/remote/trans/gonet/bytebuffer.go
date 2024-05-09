package gonet

import (
	"errors"
	"io"
	"sync"

	"github.com/yamakiller/velcro-go/rpc/pkg/remote"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/velcronet"
)

var rwPool sync.Pool

func init() {
	rwPool.New = newBufferReadWriter
}

type bufferReadWriter struct {
	reader circbuf.Reader
	writer circbuf.Writer

	ioReader io.Reader
	ioWriter io.Writer

	readSize int
	status   int
}

func newBufferReadWriter() interface{} {
	return &bufferReadWriter{}
}

// NewBufferReader 使用指定IO读取器创建一个新的 remote.ByteBuffer.
func NewBufferReader(ir io.Reader) remote.ByteBuffer {
	rw := rwPool.Get().(*bufferReadWriter)
	if npReader, ok := ir.(interface{ Reader() circbuf.Reader }); ok {
		rw.reader = npReader.Reader()
	} else {
		rw.reader = velcronet.NewReader(ir)
	}
	rw.ioReader = ir
	rw.status = remote.BitReadable
	rw.readSize = 0
	return rw
}

// NewBufferWriter 使用指定IO写入器创建一个新的 remote.ByteBuffer.
func NewBufferWriter(iw io.Writer) remote.ByteBuffer {
	rw := rwPool.Get().(*bufferReadWriter)
	rw.writer = velcronet.NewWriter(iw)
	rw.ioWriter = iw
	rw.status = remote.BitWritable
	return rw
}

// NewBufferReadWriter 使用指定IO读写器创建一个新的remote.ByteBuffer.
func NewBufferReadWriter(irw io.ReadWriter) remote.ByteBuffer {
	rw := rwPool.Get().(*bufferReadWriter)
	rw.writer = velcronet.NewWriter(irw)
	rw.reader = velcronet.NewReader(irw)
	rw.ioWriter = irw
	rw.ioReader = irw
	rw.status = remote.BitWritable | remote.BitReadable
	return rw
}

func (rw *bufferReadWriter) readable() bool {
	return rw.status&remote.BitReadable != 0
}

func (rw *bufferReadWriter) writable() bool {
	return rw.status&remote.BitWritable != 0
}

func (rw *bufferReadWriter) Next(n int) (p []byte, err error) {
	if !rw.readable() {
		return nil, errors.New("unreadable buffer, cannot support Next")
	}
	if p, err = rw.reader.Next(n); err == nil {
		rw.readSize += n
	}
	return
}

func (rw *bufferReadWriter) Peek(n int) (buf []byte, err error) {
	if !rw.readable() {
		return nil, errors.New("unreadable buffer, cannot support Peek")
	}
	return rw.reader.Peek(n)
}

func (rw *bufferReadWriter) Skip(n int) (err error) {
	if !rw.readable() {
		return errors.New("unreadable buffer, cannot support Skip")
	}
	return rw.reader.Skip(n)
}

func (rw *bufferReadWriter) ReadableLen() (n int) {
	if !rw.readable() {
		return -1
	}
	return rw.reader.Length()
}

func (rw *bufferReadWriter) ReadString(n int) (s string, err error) {
	if !rw.readable() {
		return "", errors.New("unreadable buffer, cannot support ReadString")
	}
	if s, err = rw.reader.ReadString(n); err == nil {
		rw.readSize += n
	}
	return
}

func (rw *bufferReadWriter) ReadBinary(n int) (p []byte, err error) {
	if !rw.readable() {
		return p, errors.New("unreadable buffer, cannot support ReadBinary")
	}
	if p, err = rw.reader.ReadBinary(n); err == nil {
		rw.readSize += n
	}
	return
}

func (rw *bufferReadWriter) Read(p []byte) (n int, err error) {
	if !rw.readable() {
		return -1, errors.New("unreadable buffer, cannot support Read")
	}
	if rw.ioReader != nil {
		return rw.ioReader.Read(p)
	}
	return -1, errors.New("ioReader is nil")
}

func (rw *bufferReadWriter) ReadLen() (n int) {
	return rw.readSize
}

func (rw *bufferReadWriter) Malloc(n int) (buf []byte, err error) {
	if !rw.writable() {
		return nil, errors.New("unwritable buffer, cannot support Malloc")
	}
	return rw.writer.Malloc(n)
}

func (rw *bufferReadWriter) MallocLen() (length int) {
	if !rw.writable() {
		return -1
	}
	return rw.writer.MallocLen()
}

func (rw *bufferReadWriter) WriteString(s string) (n int, err error) {
	if !rw.writable() {
		return -1, errors.New("unwritable buffer, cannot support WriteString")
	}
	return rw.writer.WriteString(s)
}

func (rw *bufferReadWriter) WriteBinary(b []byte) (n int, err error) {
	if !rw.writable() {
		return -1, errors.New("unwritable buffer, cannot support WriteBinary")
	}
	return rw.writer.WriteBinary(b)
}

func (rw *bufferReadWriter) Flush() (err error) {
	if !rw.writable() {
		return errors.New("unwritable buffer, cannot support Flush")
	}
	return rw.writer.Flush()
}

func (rw *bufferReadWriter) Write(p []byte) (n int, err error) {
	if !rw.writable() {
		return -1, errors.New("unwritable buffer, cannot support Write")
	}
	if rw.ioWriter != nil {
		return rw.ioWriter.Write(p)
	}
	return -1, errors.New("ioWriter is nil")
}

func (rw *bufferReadWriter) Release(e error) (err error) {
	if rw.reader != nil {
		err = rw.reader.Release()
	}
	rw.zero()
	rwPool.Put(rw)
	return
}

// WriteDirect 是一种不复制地写入 []byte 的方法, 并分割原始缓冲区.
func (rw *bufferReadWriter) WriteDirect(p []byte, remainCap int) error {
	if !rw.writable() {
		return errors.New("unwritable buffer, cannot support WriteBinary")
	}
	return rw.writer.WriteDirect(p, remainCap)
}

func (rw *bufferReadWriter) AppendBuffer(buf remote.ByteBuffer) (err error) {
	subBuf, ok := buf.(*bufferReadWriter)
	if !ok {
		return errors.New("AppendBuffer failed, Buffer is not bufferReadWriter")
	}
	if err = rw.writer.Append(subBuf.writer); err != nil {
		return
	}
	return buf.Release(nil)
}

// NewBuffer 返回一个新的可写的 remote.ByteBuffer.
func (rw *bufferReadWriter) NewBuffer() remote.ByteBuffer {
	panic("not implemented")
}

func (rw *bufferReadWriter) Bytes() (buf []byte, err error) {
	panic("not implemented")
}

func (rw *bufferReadWriter) zero() {
	rw.reader = nil
	rw.writer = nil
	rw.ioReader = nil
	rw.ioWriter = nil
	rw.readSize = 0
	rw.status = 0
}
