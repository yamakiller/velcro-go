package circbuf

import (
	"errors"
	"sync"
	"unsafe"
)

var (
	ErrTooManyDataToWrite = errors.New("too many data to write")
	ErrIsFull             = errors.New("ringbuffer is full")
	ErrIsEmpty            = errors.New("ringbuffer is empty")
	ErrAccuqireLock       = errors.New("no lock to accquire")
)

type RingBuffer struct {
	_buf    []byte
	_size   int
	_r      int // next position to read
	_w      int // next position to write
	_isFull bool
	_mu     sync.Locker
}

// New returns a new RingBuffer whose buffer has the given _size.
func New(size int, mutex sync.Locker) *RingBuffer {
	return &RingBuffer{
		_buf:  make([]byte, size),
		_size: size,
		_mu:   mutex,
	}
}

// Read reads up to len(p) bytes into p. It returns the number of bytes read (0 <= n <= len(p)) and any error encountered. Even if Read returns n < len(p), it may use all of p as scratch space during the call. If some data is available but not len(p) bytes, Read conventionally returns what is available instead of waiting for more.
// When Read encounters an error or end-of-file condition after successfully reading n > 0 bytes, it returns the number of bytes read. It may return the (non-nil) error from the same call or return the error (and n == 0) from a subsequent call.
// Callers should always process the n > 0 bytes returned before considering the error err. Doing so correctly handles I/O errors that happen after reading some bytes and also both of the allowed EOF behaviors.
func (r *RingBuffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	r._mu.Lock()
	n, err = r.read(p)
	r._mu.Unlock()
	return n, err
}

func (r *RingBuffer) read(p []byte) (n int, err error) {
	if r._w == r._r && !r._isFull {
		return 0, ErrIsEmpty
	}

	if r._w > r._r {
		n = r._w - r._r
		if n > len(p) {
			n = len(p)
		}
		copy(p, r._buf[r._r:r._r+n])
		r._r = (r._r + n) % r._size
		return
	}

	n = r._size - r._r + r._w
	if n > len(p) {
		n = len(p)
	}

	if r._r+n <= r._size {
		copy(p, r._buf[r._r:r._r+n])
	} else {
		c1 := r._size - r._r
		copy(p, r._buf[r._r:r._size])
		c2 := n - c1
		copy(p[c1:], r._buf[0:c2])
	}
	r._r = (r._r + n) % r._size

	r._isFull = false

	return n, err
}

// ReadByte reads and returns the next byte from the input or ErrIsEmpty.
func (r *RingBuffer) ReadByte() (b byte, err error) {
	r._mu.Lock()
	if r._w == r._r && !r._isFull {
		r._mu.Unlock()
		return 0, ErrIsEmpty
	}
	b = r._buf[r._r]
	r._r++
	if r._r == r._size {
		r._r = 0
	}

	r._isFull = false
	r._mu.Unlock()
	return b, err
}

// Write writes len(p) bytes from p to the underlying _buf.
// It returns the number of bytes written from p (0 <= n <= len(p)) and any error encountered that caused the write to stop early.
// Write returns a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
func (r *RingBuffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	r._mu.Lock()
	n, err = r.write(p)
	r._mu.Unlock()

	return n, err
}

func (r *RingBuffer) write(p []byte) (n int, err error) {
	if r._isFull {
		return 0, ErrIsFull
	}

	var avail int
	if r._w >= r._r {
		avail = r._size - r._w + r._r
	} else {
		avail = r._r - r._w
	}

	if len(p) > avail {
		err = ErrTooManyDataToWrite
		p = p[:avail]
	}
	n = len(p)

	if r._w >= r._r {
		c1 := r._size - r._w
		if c1 >= n {
			copy(r._buf[r._w:], p)
			r._w += n
		} else {
			copy(r._buf[r._w:], p[:c1])
			c2 := n - c1
			copy(r._buf[0:], p[c1:])
			r._w = c2
		}
	} else {
		copy(r._buf[r._w:], p)
		r._w += n
	}

	if r._w == r._size {
		r._w = 0
	}
	if r._w == r._r {
		r._isFull = true
	}

	return n, err
}

// WriteByte writes one byte into buffer, and returns ErrIsFull if buffer is full.
func (r *RingBuffer) WriteByte(c byte) error {
	r._mu.Lock()
	err := r.writeByte(c)
	r._mu.Unlock()
	return err
}

func (r *RingBuffer) writeByte(c byte) error {
	if r._w == r._r && r._isFull {
		return ErrIsFull
	}
	r._buf[r._w] = c
	r._w++

	if r._w == r._size {
		r._w = 0
	}
	if r._w == r._r {
		r._isFull = true
	}

	return nil
}

// Length return the length of available read bytes.
func (r *RingBuffer) Length() int {
	r._mu.Lock()
	defer r._mu.Unlock()

	if r._w == r._r {
		if r._isFull {
			return r._size
		}
		return 0
	}

	if r._w > r._r {
		return r._w - r._r
	}

	return r._size - r._r + r._w
}

// Capacity returns the _size of the underlying buffer.
func (r *RingBuffer) Capacity() int {
	return r._size
}

// Free returns the length of available bytes to write.
func (r *RingBuffer) Free() int {
	r._mu.Lock()
	defer r._mu.Unlock()

	if r._w == r._r {
		if r._isFull {
			return 0
		}
		return r._size
	}

	if r._w < r._r {
		return r._r - r._w
	}

	return r._size - r._w + r._r
}

// WriteString writes the contents of the string s to buffer, which accepts a slice of bytes.
func (r *RingBuffer) WriteString(s string) (n int, err error) {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	_buf := *(*[]byte)(unsafe.Pointer(&h))
	return r.Write(_buf)
}

// Bytes returns all available read bytes. It does not move the read pointer and only copy the available data.
func (r *RingBuffer) Bytes() []byte {
	r._mu.Lock()
	defer r._mu.Unlock()

	if r._w == r._r {
		if r._isFull {
			_buf := make([]byte, r._size)
			copy(_buf, r._buf[r._r:])
			copy(_buf[r._size-r._r:], r._buf[:r._w])
			return _buf
		}
		return nil
	}

	if r._w > r._r {
		_buf := make([]byte, r._w-r._r)
		copy(_buf, r._buf[r._r:r._w])
		return _buf
	}

	n := r._size - r._r + r._w
	_buf := make([]byte, n)

	if r._r+n < r._size {
		copy(_buf, r._buf[r._r:r._r+n])
	} else {
		c1 := r._size - r._r
		copy(_buf, r._buf[r._r:r._size])
		c2 := n - c1
		copy(_buf[c1:], r._buf[0:c2])
	}

	return _buf
}

// IsFull returns this ringbuffer is full.
func (r *RingBuffer) IsFull() bool {
	r._mu.Lock()
	defer r._mu.Unlock()

	return r._isFull
}

// IsEmpty returns this ringbuffer is empty.
func (r *RingBuffer) IsEmpty() bool {
	r._mu.Lock()
	defer r._mu.Unlock()

	return !r._isFull && r._w == r._r
}

// Reset the read pointer and writer pointer to zero.
func (r *RingBuffer) Reset() {
	r._mu.Lock()
	defer r._mu.Unlock()

	r._r = 0
	r._w = 0
	r._isFull = false
}
