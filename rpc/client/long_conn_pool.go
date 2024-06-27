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
		idles:        intrusive.NewLinked(&syncx.NoMutex{}),
		busys:        intrusive.NewLinked(&syncx.NoMutex{}),
		stopper:      make(chan struct{}),
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

	Open()

	Close() error
}

type DefaultConnPool struct {
	idles *intrusive.Linked
	busys *intrusive.Linked

	openingConns int32
	address      string
	cfg          *LongConnPoolConfig
	stopper      chan struct{}
	mutex        sync.Mutex
}

func (dc *DefaultConnPool) Open() {
	for i := 0; i < int(dc.cfg.MaxIdleConn); i++ {
		conn := NewLongConn(dc, int64(dc.cfg.MaxIdleConnTimeout.Milliseconds()))
		err := conn.Dial(dc.address, dc.cfg.ConnectTimeout)
		if err != nil {
			continue
		}

		dc.mutex.Lock()
		dc.idles.Push(conn)
		atomic.StoreInt32(&conn.direct, LCS_Idle)
		atomic.AddInt32(&dc.openingConns, 1)
		dc.mutex.Unlock()
	}
}

func (dc *DefaultConnPool) RequestMessage(msg proto.Message, timeout int64) (proto.Message, error) {
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
	if dc.idles == nil {
		return nil, errs.ErrorRpcConnectPoolNoInitialize
	}

	dc.mutex.Lock()
	if dc.isStopped() {
		dc.mutex.Unlock()
		return nil, errs.ErrorRpcConnPoolClosed
	}

	if dc.idles.Len() == 0 && dc.openingConns >= dc.cfg.MaxConn {
		dc.mutex.Unlock()
		return nil, errs.ErrorRpcConnectMaxmize
	}

	if dc.idles.Len() == 0 {
		atomic.AddInt32(&dc.openingConns, 1)
		dc.mutex.Unlock()

		conn := NewLongConn(dc, int64(dc.cfg.MaxIdleConnTimeout.Milliseconds()))
		err := conn.Dial(dc.address, dc.cfg.ConnectTimeout)
		dc.mutex.Lock()
		if err != nil {
			atomic.AddInt32(&dc.openingConns, -1)
			dc.mutex.Unlock()
			return nil, err
		}

		if !conn.IsConnected() {
			atomic.AddInt32(&dc.openingConns, -1)
			dc.mutex.Unlock()
			return nil, errs.ErrorRpcConnectorClosed
		}

		atomic.StoreInt32(&conn.direct, LCS_Busy)
		dc.busys.Push(conn)
		dc.mutex.Unlock()

		atomic.StoreInt64(&conn.usedLastTime, time.Now().UnixMilli())
		return conn, nil
	}

	select {
	case <-ctx.Done():
		dc.mutex.Unlock()
		return nil, errs.ErrorRequestTimeout
	default:
		node := dc.idles.Pop()
		dc.busys.Push(node)
		conn := node.(*LongConn)
		atomic.StoreInt32(&conn.direct, LCS_Busy)
		dc.mutex.Unlock()

		atomic.StoreInt64(&conn.usedLastTime, time.Now().UnixMilli())
		return conn, nil
	}
}

func (dc *DefaultConnPool) Put(conn *LongConn) error {
	if conn == nil {
		return errs.ErrorInvalidRpcConnect
	}
	if dc.idles == nil {
		return errs.ErrorRpcConnectPoolNoInitialize
	}

	if !conn.IsConnected() {
		err := dc.Discard(conn)
		conn = nil
		return err
	}

	dc.mutex.Lock()
	if atomic.LoadInt32(&conn.direct) != LCS_Unknown {
		dc.busys.Remove(conn)
		dc.idles.Push(conn)
		atomic.StoreInt32(&conn.direct, LCS_Idle)
	}
	dc.mutex.Unlock()
	return nil
}

func (dc *DefaultConnPool) Discard(conn *LongConn) error {
	dc.mutex.Lock()
	if atomic.LoadInt32(&conn.direct) == LCS_Busy {
		dc.busys.Remove(conn)
		atomic.AddInt32(&dc.openingConns, -1)
		atomic.StoreInt32(&conn.direct, LCS_Unknown)
	} else if atomic.LoadInt32(&conn.direct) == LCS_Idle {
		dc.idles.Remove(conn)
		atomic.AddInt32(&dc.openingConns, -1)
		atomic.StoreInt32(&conn.direct, LCS_Unknown)
	}
	dc.mutex.Unlock()
	return nil
}

func (dc *DefaultConnPool) Close() error {
	dc.mutex.Lock()
	close(dc.stopper)
	dc.idles.Foreach(func(node intrusive.INode) bool {
		node.(*LongConn).Close()
		return true
	})
	dc.busys.Foreach(func(node intrusive.INode) bool {
		node.(*LongConn).Close()
		return true
	})
	
	return nil
}

func (dc *DefaultConnPool) Tick() {
	dc.mutex.Lock()
	dc.idles.Foreach(func(node intrusive.INode) bool {
		if node == nil {
			return false
		}

		if atomic.LoadInt32(&dc.openingConns) <= dc.cfg.MaxIdleConn {
			return true
		}

		if !node.(*LongConn).IsConnected() {
			return true
		}

		usedLastTime := atomic.LoadInt64(&node.(*LongConn).usedLastTime)
		if time.Now().UnixMilli()-usedLastTime > int64(dc.cfg.MaxIdleConnTimeout.Milliseconds()) {
			node.(*LongConn).Close()
		}

		return true
	})
	dc.mutex.Unlock()
}

func (dc *DefaultConnPool) isStopped() bool {
	select {
	case <-dc.stopper:
		return true
	default:
		return false
	}
}
