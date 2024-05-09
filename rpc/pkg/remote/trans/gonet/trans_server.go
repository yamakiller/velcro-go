package gonet

import (
	"context"
	"errors"
	"net"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/yamakiller/velcro-go/rpc/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/syncx"
	"github.com/yamakiller/velcro-go/utils/velcronet"
	"github.com/yamakiller/velcro-go/vlog"
)

type transServer struct {
	opt          *remote.ServerOption
	transHdlr    remote.ServerTransHandler
	listener     net.Listener
	listenerCfg  net.ListenConfig
	connOfNumber syncx.AtomicInt
	sync.Mutex
}

var _ remote.TransServer = &transServer{}

func (ts *transServer) CreateListener(addr net.Addr) (ln net.Listener, err error) {
	ln, err = ts.listenerCfg.Listen(context.Background(), addr.Network(), addr.String())
	return ln, err
}

// BootstrapServer ...
func (ts *transServer) BootstrapServer(ln net.Listener) error {
	if ln == nil {
		return errors.New("listener is nil in gonet transport server")
	}

	ts.listener = ln

	for {
		conn, err := ts.listener.Accept()
		if err != nil {
			vlog.Errorf("VELCRO: BootstrapServer accept failed, err=%s", err.Error())
			os.Exit(1)
		}

		go func() {
			var (
				ctx = context.Background()
				err error
			)
			defer func() {
				transRecover(ctx, conn, "OnRead")
			}()

			ibc := newIOBufferConnector(conn)
			ctx, err = ts.transHdlr.OnActive(ctx, ibc)
			if err != nil {
				vlog.ContextErrorf(ctx, "VELCRO: OnActive error=%s", err)
				return
			}

			for {
				ts.refreshDeadline(rpcinfo.GetRPCInfo(ctx), ibc)
				err := ts.transHdlr.OnRead(ctx, ibc)
				if err != nil {
					ts.onError(ctx, err, ibc)
					_ = ibc.Close()
					return
				}
			}
		}()
	}
}

func (ts *transServer) Shutdown() error {
	ts.Lock()
	defer ts.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), ts.opt.ExitWaitTime)
	defer cancel()
	if g, ok := ts.transHdlr.(remote.GracefulShutdown); ok {
		if ts.listener != nil {
			_ = ts.listener.Close()
			_ = g.GracefulShutdown(ctx)
		}
	}
	return nil
}

func (ts *transServer) ConnectorOfNumber() syncx.AtomicInt {
	return ts.connOfNumber
}

func (ts *transServer) onError(ctx context.Context, err error, conn net.Conn) {
	ts.transHdlr.OnError(ctx, err, conn)
}

func (ts *transServer) refreshDeadline(ri rpcinfo.RPCInfo, conn net.Conn) {
	readTimeout := ri.Config().ReadWriteTimeout()
	_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
}

func newIOBufferConnector(c net.Conn) *ioBufferConnector {
	return &ioBufferConnector{
		c: c,
		r: velcronet.NewReader(c),
	}
}

type ioBufferConnector struct {
	c net.Conn
	r circbuf.Reader
}

func (ibc *ioBufferConnector) RawConn() net.Conn {
	return ibc.c
}

func (ibc *ioBufferConnector) Read(b []byte) (int, error) {
	buf, err := ibc.r.Next(len(b))
	if err != nil {
		return 0, err
	}
	copy(b, buf)
	return len(b), nil
}

func (ibc *ioBufferConnector) Write(b []byte) (int, error) {
	return ibc.c.Write(b)
}

func (ibc *ioBufferConnector) Close() error {
	ibc.r.Release()
	return ibc.c.Close()
}

func (ibc *ioBufferConnector) LocalAddr() net.Addr {
	return ibc.c.LocalAddr()
}

func (ibc *ioBufferConnector) RemoteAddr() net.Addr {
	return ibc.c.RemoteAddr()
}

func (ibc *ioBufferConnector) SetDeadline(t time.Time) error {
	return ibc.c.SetDeadline(t)
}

func (ibc *ioBufferConnector) SetReadDeadline(t time.Time) error {
	return ibc.c.SetReadDeadline(t)
}

func (ibc *ioBufferConnector) SetWriteDeadline(t time.Time) error {
	return ibc.c.SetWriteDeadline(t)
}

func (ibc *ioBufferConnector) Reader() circbuf.Reader {
	return ibc.r
}

func transRecover(ctx context.Context, c net.Conn, funcName string) {
	panicErr := recover()
	if panicErr != nil {
		if c != nil {
			vlog.ContextErrorf(ctx, "VELCRO: panic happened in %s, remoteAddress=%s, error=%v\nstack=%s",
				funcName, c.RemoteAddr(), panicErr, string(debug.Stack()))
			return
		}

		vlog.ContextErrorf(ctx, "VELCRO: panic happened in %s, error=%v\nstack=%s",
			funcName, panicErr, string(debug.Stack()))
	}
}
