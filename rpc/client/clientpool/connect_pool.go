package clientpool

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/rpc/client"
	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
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

func (cp *ConnectPool) RequestMessage(msg protoreflect.ProtoMessage, timeout int64) (client.IFuture, error) {
	var (
		conn client.IConnect
		res  client.IFuture
		err  error
	)
	conn, err = cp.Get()
	if err != nil {
		return nil, err
	}
	res, err = conn.RequestMessage(msg, timeout)
	if err == errs.ErrorRpcConnectorClosed{
		cp.Remove(conn.Node())
	}else{
		cp.Put(conn)
	}
	
	return res, err
}

func (cp *ConnectPool) Tick() {
	remove_list := make([]intrusive.INode, 0)
	cp.pls.Foreach(func(node intrusive.INode) bool {
		if node == nil {
			return false
		}
		if node.(*client.PoolLinkNode).Conn.Timeout()-time.Now().UnixMilli() <= 0 {
			remove_list = append(remove_list, node)
		}
		return true
	})

	for _, node := range remove_list {
		cp.Close(node.(*client.PoolLinkNode).Conn)
		cp.Remove(node)
	}
}

func (cp *ConnectPool) Get() (client.IConnect, error) {

	node := cp.pls.Pop()

	if node != nil {
		conn := node.(*client.PoolLinkNode).Conn
		conn.WithTimeout(time.Now().Add(time.Duration(cp.config.MaxIdleConnTimeout)).UnixMilli())
		return conn, nil
	} else {
		conn :=cp.config.NewConn(
			client.WithConnected(cp.config.Connected),
			client.WithClosed(cp.config.Closed),
			client.WithKleepalive(cp.config.Kleepalive),
		)
		
		err := conn.Dial(cp.address, cp.config.MaxIdleConnTimeout)
		if err != nil {
			return nil, err
		}
		node = &client.PoolLinkNode{Conn: conn}
		conn.WithAffiliation(cp)
		conn.WithTimeout(time.Now().Add(time.Duration(cp.config.MaxIdleConnTimeout)).UnixMilli())
		conn.WithNode(node)
		atomic.AddInt32(&cp.openingConns, 1)
		fmt.Fprintf(os.Stderr, "ConnectPool ++ %d\n", cp.openingConns)
		cp.pls.Push(node)
		return conn, nil
	}
}

func (cp *ConnectPool) Put(conn client.IConnect) error {

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
	fmt.Fprintf(os.Stderr, "ConnectPool -- %d\n", cp.openingConns)
}

func (cp *ConnectPool) Close(conn client.IConnect) error {
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
