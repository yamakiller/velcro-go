package velcronet

import (
	"fmt"
	"net"
	"syscall"
)

// 扩展系统错误
const (
	// 连接已经关闭.
	ErrConnClosed = syscall.Errno(0x101)
	// 读取I/O缓冲区超时, 在调用 连接的Reader时.
	ErrReadTimeout = syscall.Errno(0x102)
	// 连接超时
	ErrDialTimeout = syscall.Errno(0x103)
	// 调用 dialer 时超时
	ErrDialNoDeadline = syscall.Errno(0x104)
	// 调用会拿书不支持
	ErrUnsupported = syscall.Errno(0x105)
	// == io.EOF
	ErrEOF = syscall.Errno(0x106)
	// 写入I/O缓冲区超时, 在调用 连接的Writer时.
	ErrWriteTimeout = syscall.Errno(0x107)
)

const ErrnoMask = 0xFF

func Exception(err error, suffix string) error {
	var code, ok = err.(syscall.Errno)
	if !ok {
		if suffix == "" {
			return err
		}
		return fmt.Errorf("%s %s", err.Error(), suffix)
	}
	return &exception{code: code, suffix: suffix}
}

var (
	_ net.Error = (*exception)(nil)
)

type exception struct {
	code   syscall.Errno
	suffix string
}

func (e *exception) Error() string {
	var s string
	if int(e.code)&0x100 != 0 {
		s = errnos[int(e.code)&ErrnoMask]
	}
	if s == "" {
		s = e.code.Error()
	}
	if e.suffix != "" {
		s += " " + e.suffix
	}
	return s
}

func (e *exception) Is(target error) bool {
	if e == target {
		return true
	}
	if e.code == target {
		return true
	}
	// TODO: ErrConnClosed contains ErrEOF
	if e.code == ErrEOF && target == ErrConnClosed {
		return true
	}
	return e.code.Is(target)
}

func (e *exception) Unwrap() error {
	return e.code
}

func (e *exception) Timeout() bool {
	switch e.code {
	case ErrDialTimeout, ErrReadTimeout, ErrWriteTimeout:
		return true
	}
	if e.code.Timeout() {
		return true
	}
	return false
}

func (e *exception) Temporary() bool {
	return e.code.Temporary()
}

var errnos = [...]string{
	ErrnoMask & ErrConnClosed:     "connection has been closed",
	ErrnoMask & ErrReadTimeout:    "connection read timeout",
	ErrnoMask & ErrDialTimeout:    "dial wait timeout",
	ErrnoMask & ErrDialNoDeadline: "dial no deadline",
	ErrnoMask & ErrUnsupported:    "netpoll dose not support",
	ErrnoMask & ErrEOF:            "EOF",
	ErrnoMask & ErrWriteTimeout:   "connection write timeout",
}
