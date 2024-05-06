package velcronet

import (
	"fmt"
	"io"

	"github.com/yamakiller/velcro-go/utils/circbuf"
)

var _ circbuf.Reader = &vReader{}
var _ circbuf.Writer = &vWriter{}

const _MaxReadCycle = 16

func newVReader(r io.Reader) *vReader {
	return &vReader{
		r: r,
		b: circbuf.NewLinkBuffer(),
	}
}

func newVWriter(w io.Writer) *vWriter {
	return &vWriter{
		w: w,
		b: circbuf.NewLinkBuffer(),
	}
}

func newIOReader(r circbuf.Reader) *ioReader {
	return &ioReader{
		r: r,
	}
}

func newIOWriter(w circbuf.Writer) *ioWriter {
	return &ioWriter{
		w: w,
	}
}

type ioReader struct {
	r circbuf.Reader
}

func (r *ioReader) Read(p []byte) (n int, err error) {
	l := len(p)
	if l == 0 {
		return 0, nil
	}
	// read min(len(p), buffer.Len)
	if has := r.r.Length(); has < l {
		l = has
	}
	if l == 0 {
		return 0, io.EOF
	}
	src, err := r.r.Next(l)
	if err != nil {
		return 0, err
	}
	n = copy(p, src)
	err = r.r.Release()
	if err != nil {
		return 0, err
	}
	return n, nil
}

type ioWriter struct {
	w circbuf.Writer
}

func (w *ioWriter) Write(p []byte) (n int, err error) {
	dst, err := w.w.Malloc(len(p))
	if err != nil {
		return 0, err
	}
	n = copy(dst, p)
	err = w.w.Flush()
	if err != nil {
		return 0, err
	}
	return n, nil
}

// vReader 实现.
type vReader struct {
	r io.Reader
	b *circbuf.LinkBuffer
}

// Next ...
func (r *vReader) Next(n int) ([]byte, error) {
	if err := r.waitRead(n); err != nil {
		return nil, err
	}
	return r.b.Next(n)
}

// Peek ...
func (r *vReader) Peek(n int) ([]byte, error) {
	if err := r.waitRead(n); err != nil {
		return nil, err
	}

	return r.b.Peek(n)
}

// Skip ...
func (r *vReader) Skip(n int) error {
	if err := r.waitRead(n); err != nil {
		return err
	}

	return r.b.Skip(n)
}

// Release ...
func (r *vReader) Release() error {
	return r.b.Release()
}

// Slice ...
func (r *vReader) Slice(n int) (circbuf.Reader, error) {
	if err := r.waitRead(n); err != nil {
		return nil, err
	}

	return r.b.Slice(n)
}

// Length ...
func (r *vReader) Length() int {
	return r.b.Length()
}

// ReadString ...
func (r *vReader) ReadString(n int) (s string, err error) {
	if err = r.waitRead(n); err != nil {
		return s, err
	}

	return r.b.ReadString(n)
}

// ReadBinary ...
func (r *vReader) ReadBinary(n int) (p []byte, err error) {
	if err = r.waitRead(n); err != nil {
		return p, err
	}

	return r.b.ReadBinary(n)
}

// ReadByte ...
func (r *vReader) ReadByte() (b byte, err error) {
	if err = r.waitRead(1); err != nil {
		return b, err
	}

	return r.b.ReadByte()
}

// Until ...
func (r *vReader) Until(delim byte) (line []byte, err error) {
	return r.b.Until(delim)
}

func (r *vReader) waitRead(n int) (err error) {
	for r.b.Length() < n {
		err = r.fill(n)
		if err != nil {
			if err == io.EOF {
				err = Exception(ErrEOF, "")
			}
			return err
		}
	}
	return nil
}

func (r *vReader) fill(n int) (err error) {
	var buf []byte
	var num int
	for i := 0; i < _MaxReadCycle && r.b.Length() < n && err == nil; i++ {
		buf, err = r.b.Malloc(circbuf.LinkBufferCap)
		if err != nil {
			return err
		}
		num, err = r.r.Read(buf)
		if num < 0 {
			if err == nil {
				err = fmt.Errorf("vReader fill negative count[%d]", num)
			}
			num = 0
		}
		r.b.MallocAck(num)
		r.b.Flush()
		if err != nil {
			return err
		}
	}
	return err
}

// vWriter 实现 Writer
type vWriter struct {
	w io.Writer
	b *circbuf.LinkBuffer
}

// Malloc ...
func (w *vWriter) Malloc(n int) ([]byte, error) {
	return w.b.Malloc(n)
}

// MallocLen ...
func (w *vWriter) MallocLen() int {
	return w.b.MallocLen()
}

// Flush ...
func (w *vWriter) Flush() error {
	w.b.Flush()
	n, err := w.w.Write(w.b.Bytes())
	if n > 0 {
		w.b.Skip(n)
		w.b.Release()
	}

	return err
}

// MallocAck ...
func (w *vWriter) MallocAck(n int) error {
	return w.b.MallocAck(n)
}

// Append ...
func (w *vWriter) Append(w2 circbuf.Writer) error {
	return w.b.Append(w2)
}

// WriteString ...
func (w *vWriter) WriteString(s string) (int, error) {
	return w.b.WriteString(s)
}

// WriteBinary ...
func (w *vWriter) WriteBinary(b []byte) (int, error) {
	return w.b.WriteBinary(b)
}

// WriteDirect ...
func (w *vWriter) WriteDirect(p []byte, remainCap int) error {
	return w.b.WriteDirect(p, remainCap)
}

// WriteByte ...
func (w *vWriter) WriteByte(b byte) error {
	return w.b.WriteByte(b)
}

// vWriter 实现 ReadWriter.
type vReadWriter struct {
	*vReader
	*vWriter
}

// ioWriter 实现 ioReadWriter
type ioReadWriter struct {
	*ioReader
	*ioWriter
}
