package client

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
	"github.com/yamakiller/velcro-go/utils/syncx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	sharedTickers sync.Map
)

func getSharedTicker(p *DefaultConnPool, refreshInterval time.Duration) *utils.SharedTicker {
	sti, ok := sharedTickers.Load(refreshInterval)
	if ok {
		st := sti.(*utils.SharedTicker)
		st.Add(p)
		return st
	}
	sti, _ = sharedTickers.LoadOrStore(refreshInterval, utils.NewSharedTicker(refreshInterval))
	st := sti.(*utils.SharedTicker)
	st.Add(p)
	return st
}

func NewDefaultConnPool(address string, idlConfig LongConnPoolConfig) *DefaultConnPool {

	res := &DefaultConnPool{
		openingConns: 0,
		pls:          intrusive.NewLinked(&syncx.NoMutex{}),
	}
	res.address = address
	res.cfg = CheckLongConnPoolConfig(idlConfig)
	getSharedTicker(res, time.Duration(res.cfg.MaxIdleConnTimeout))
	return res
}

type LongConnPool interface {
	RequestMessage(protoreflect.ProtoMessage, int64) (proto.Message, error)

	Get(ctx context.Context, address string) (*LongConn, error)

	Put(*LongConn) error

	Discard(*LongConn) error

	Close() error
}

type DefaultConnPool struct {
	pls          *intrusive.Linked
	openingConns int32
	address      string
	cfg          *LongConnPoolConfig
	mutex        sync.Mutex
}

func (dc *DefaultConnPool) RequestMessage(msg protoreflect.ProtoMessage, timeout int64) (proto.Message, error) {
	var (
		conn *LongConn
		res  proto.Message
		err  error
		ctx  context.Context
	)

	if timeout > 0 {
		ctx, _ = context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	} else {
		ctx = context.Background()
	}
	conn, err = dc.Get(ctx, dc.address)
	if err != nil {
		return nil, err
	}
	defer dc.Put(conn)
	res, err = conn.RequestMessage(msg, timeout)

	return res, err
}

func (dc *DefaultConnPool) Get(ctx context.Context, address string) (*LongConn, error) {
	if dc.pls == nil {
		return nil, errs.ErrorRpcConnectPoolNoInitialize
	}

	dc.mutex.Lock()
	if dc.pls.Len() == 0 && dc.openingConns >= dc.cfg.MaxConn {
		dc.mutex.Unlock()
		return nil, errs.ErrorRpcConnectMaxmize
	}

	for {
		if dc.pls.Len() == 0 {

			conn := NewLongConn(dc, dc.cfg.MaxIdleConnTimeout.Milliseconds())
			err := conn.Dial(address, 2000)
			if err != nil {
				continue
			}

			if !conn.IsConnected() {
				continue
			}

			atomic.AddInt32(&dc.openingConns, 1)
			dc.mutex.Unlock()
			return conn, nil
		}

		select {
		case <-ctx.Done():
			dc.mutex.Unlock()
			return nil, errs.ErrorRequestTimeout
		default:
			node := dc.pls.Pop()
			conn := node.(*LongConn)

			if !conn.IsConnected() {
				if atomic.LoadInt32(&dc.openingConns) >= dc.cfg.MaxIdleConn {
					atomic.AddInt32(&dc.openingConns, -1)
				}
				continue
			}
			dc.mutex.Unlock()
			return conn, nil
		}
	}

}

func (dc *DefaultConnPool) Put(conn *LongConn) error {
	if conn == nil {
		return errs.ErrorInvalidRpcConnect
	}
	if dc.pls == nil {
		return errs.ErrorRpcConnectPoolNoInitialize
	}

	if atomic.LoadInt32(&dc.openingConns) > dc.cfg.MaxConn || !conn.IsConnected() {
		atomic.AddInt32(&dc.openingConns, -1)
		err := dc.Discard(conn)
		conn = nil
		return err
	}

	if !conn.IsConnected() {
		atomic.AddInt32(&dc.openingConns, -1)
		err := dc.Discard(conn)
		conn = nil
		return err
	}
	dc.mutex.Lock()
	dc.pls.Push(conn)
	dc.mutex.Unlock()
	return nil
}

func (dc *DefaultConnPool) Discard(conn *LongConn) error {
	return nil
}

func (dc *DefaultConnPool) Close() error {
	dc.mutex.Lock()
	dc.pls.Foreach(func(node intrusive.INode) bool {
		node.(*LongConn).Close()
		return true
	})
	dc.pls = nil
	dc.mutex.Unlock()
	return nil
}

func (dc *DefaultConnPool) Tick() {
	dc.mutex.Lock()
	dc.pls.Foreach(func(node intrusive.INode) bool {
		if node == nil {
			return false
		}
		if time.Now().UnixMilli()-node.(*LongConn).usedLastTime < 0 {
			return true
		}

		if atomic.LoadInt32(&dc.openingConns) >= dc.cfg.MaxIdleConn {
			node.(*LongConn).Close()
			dc.pls.Remove(node)
			atomic.AddInt32(&dc.openingConns, -1)
		}
		return true
	})
	dc.mutex.Unlock()
}
