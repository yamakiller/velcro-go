package velcronet

import (
	"io"

	"github.com/yamakiller/velcro-go/utils/circbuf"
)

// NewReader 创建一个虚拟缓冲读取器
func NewReader(r io.Reader) circbuf.Reader {
	return newVReader(r)
}

// NewWriter 创建一个虚拟缓冲写入器
func NewWriter(w io.Writer) circbuf.Writer {
	return newVWriter(w)
}

// NewReadWriter 创建一个虚拟缓冲读写器
func NewReadWriter(rw io.ReadWriter) circbuf.ReadWriter {
	return &vReadWriter{
		vReader: newVReader(rw),
		vWriter: newVWriter(rw),
	}
}

func ToIOReader(r circbuf.Reader) io.Reader {
	if reader, ok := r.(io.Reader); ok {
		return reader
	}
	return newIOReader(r)
}

func ToIOWriter(w circbuf.Writer) io.Writer {
	if writer, ok := w.(io.Writer); ok {
		return writer
	}
	return newIOWriter(w)
}

func ToIOReadWriter(rw circbuf.ReadWriter) io.ReadWriter {
	if rwer, ok := rw.(io.ReadWriter); ok {
		return rwer
	}

	return &ioReadWriter{
		ioReader: newIOReader(rw),
		ioWriter: newIOWriter(rw),
	}
}
