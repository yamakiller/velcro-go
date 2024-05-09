package gonet

import (
	"errors"
	"io"
	"net"
	"strings"
	"syscall"

	"github.com/yamakiller/velcro-go/rpc/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc/pkg/remote/trans"
	"github.com/yamakiller/velcro-go/rpc/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/utils/velcronet"
	"golang.org/x/net/context"
)

func NewGonetExtAddon() trans.ExtAddon {
	return &gonetConnExtAddon{}
}

type gonetConnExtAddon struct{}

func (e *gonetConnExtAddon) SetReadTimeout(ctx context.Context, c net.Conn, cfg rpcinfo.RPCConfig, role remote.RPCRole) {
	//
}

func (e *gonetConnExtAddon) NewWriteByteBuffer(ctx context.Context, c net.Conn, msg remote.Message) remote.ByteBuffer {
	return NewBufferWriter(c)
}

func (e *gonetConnExtAddon) NewReadByteBuffer(ctx context.Context, c net.Conn, msg remote.Message) remote.ByteBuffer {
	return NewBufferReader(c)
}

func (e *gonetConnExtAddon) ReleaseBuffer(buffer remote.ByteBuffer, err error) error {
	if buffer != nil {
		return buffer.Release(err)
	}
	return nil
}

func (e *gonetConnExtAddon) IsTimeoutErr(err error) bool {
	if err != nil {
		if nerr, ok := err.(net.Error); ok {
			return nerr.Timeout()
		}
	}
	return false
}

func (e *gonetConnExtAddon) IsRemoteClosedErr(err error) bool {
	if err == nil {
		return false
	}
	// strings.Contains(err.Error(), "closed network connection") change to errors.Is(err, net.ErrClosed)
	// when support go version >= 1.16
	return errors.Is(err, velcronet.ErrConnClosed) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, syscall.EPIPE) ||
		strings.Contains(err.Error(), "closed network connection")
}
