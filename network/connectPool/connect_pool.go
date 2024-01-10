package connectPool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	sharedTickers sync.Map
)

func getSharedTicker(p *ConnectPool, refreshInterval time.Duration) *utils.SharedTicker {
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

func NewConnectPool(address string, idlConfig IdleConfig) *ConnectPool {

	res := &ConnectPool{
		openingConns: 0,
		pls:          intrusive.NewLinked(&sync.RWMutex{}),
	}
	res.address = address
	res.config = CheckPoolConfig(idlConfig)
	getSharedTicker(res, res.config.MaxIdleTimeout)
	return res
}

type ConnectPool struct {
	pls          *intrusive.Linked
	openingConns int32
	address      string
	config       *IdleConfig
}

func (cp *ConnectPool) RequestMessage(msg protoreflect.ProtoMessage, timeout int64) (IFuture, error) {
	var (
		conn IConnect
		res  IFuture
		err  error
	)
	conn, err = cp.Get()
	if err != nil {
		return nil, err
	}
	res, err = conn.RequestMessage(msg, timeout)
	if err == errs.ErrorRpcConnectorClosed{
		cp.Remove(conn.Node())
	}else if err ==nil{
		cp.Put(conn)
	}else{
		vlog.Errorf("PROGRAM","ConnectPool error: %v", err)
	}
	
	return res, err
}

func (cp *ConnectPool) Tick() {
	remove_list := make([]intrusive.INode, 0)
	cp.pls.Foreach(func(node intrusive.INode) bool {
		if node == nil {
			return false
		}
		if node.(*PoolLinkNode).Conn.Timeout()-time.Now().UnixMilli() <= 0 {
			remove_list = append(remove_list, node)
		}
		return true
	})

	for _, node := range remove_list {
		cp.Close(node.(*PoolLinkNode).Conn)
		cp.Remove(node)
	}
}

func (cp *ConnectPool) Get() (IConnect, error) {

	node := cp.pls.Pop()

	if node != nil {
		conn := node.(*PoolLinkNode).Conn
		conn.WithTimeout(time.Now().Add(time.Duration(cp.config.MaxIdleConnTimeout)).UnixMilli())
		return conn, nil
	} else {
		conn :=cp.config.NewConn(
			WithConnected(cp.config.Connected),
			WithClosed(cp.config.Closed),
			WithKleepalive(cp.config.Kleepalive),
		)
		
		err := conn.Dial(cp.address, cp.config.MaxIdleConnTimeout)
		if err != nil {
			return nil, err
		}
		node = &PoolLinkNode{Conn: conn}
		conn.WithAffiliation(cp)
		conn.WithTimeout(time.Now().Add(time.Duration(cp.config.MaxIdleConnTimeout)).UnixMilli())
		conn.WithNode(node)
		atomic.AddInt32(&cp.openingConns, 1)
		return conn, nil
	}
}

func (cp *ConnectPool) Put(conn IConnect) error {

	// 超过最大空闲连接数，关闭连接
	if cp.openingConns >= cp.config.MaxIdleGlobal {
		cp.Close(conn)
		cp.Remove(conn.Node())
		return nil
	}
	cp.pls.Push(conn.Node())
	return nil
}

func (cp *ConnectPool) Len() int32 {
	return cp.openingConns
}

func (cp *ConnectPool) Remove(node intrusive.INode) {
	cp.pls.Remove(node)
	atomic.AddInt32(&cp.openingConns, -1)
}

func (cp *ConnectPool) Close(conn IConnect) error {
	if conn == nil {
		return errors.New("connection is nil")
	}
	conn.Close()
	return nil
}
func (cp *ConnectPool) Shudown() {
	for {
		conn, err := cp.Get()
		if err != nil {
			return
		}
		cp.Close(conn)
	}
}
