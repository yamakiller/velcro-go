package clientpool

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/yamakiller/velcro-go/rpc/client"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
	"github.com/yamakiller/velcro-go/utils/syncx"
)



func getSharedTicker1(p *ConnectPool, refreshInterval time.Duration) *utils.SharedTicker {
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

func NewConnectPool(serviceName string, connType interface{}, idlConfig IdleConfig) *ConnectPool {
	res := &ConnectPool{
		connType:     reflect.TypeOf(connType),
		openingConns: 0,
		maxIdleConn:  idlConfig.MaxIdlePerAddress,
		pls:          intrusive.NewLinked(&syncx.NoMutex{}),
		plscon:       sync.NewCond(&sync.Mutex{}),
	}
	res.address = serviceName
	res.config = idlConfig
	getSharedTicker1(res, idlConfig.MaxIdleTimeout)
	return res
}



type ConnectPool struct {
	pls          *intrusive.Linked
	plscon       *sync.Cond
	openingConns int
	maxIdleConn  int
	connType     reflect.Type
	address string
	config  IdleConfig
}

func (cp *ConnectPool) RequestMessage(msg proto.Message) (*client.Future, error) {
	var (
		conn client.IConnect
		res *client.Future
		err error
	)
	conn, err = cp.Get()
	if err != nil {
		return nil, err
	}
	res, err = conn.RequestMessage(msg,cp.config.MaxMessageTimeout.Milliseconds())
	if err == errors.New("closed") {
		cp.Close(conn)
	} else {
		cp.Put(conn)
	}
	return res,nil
}



func (cp *ConnectPool) Tick() {
	cp.plscon.L.Lock()
	defer func() {
		cp.plscon.L.Unlock()
		cp.plscon.Signal()
	}()
	remove_list := make([]intrusive.INode, 0)
	cp.pls.Foreach(func(node intrusive.INode) bool {
		if node == nil {
			return false
		}
		if node.(client.IConnect).Timeout()-time.Now().UnixMilli() <= 0 {
			remove_list = append(remove_list, node)
		}
		return true
	})

	for _, node := range remove_list {
		cp.Close(node.(*client.PoolLinkNode).Conn)
	}
}


func (cp *ConnectPool) Get() (client.IConnect, error) {
	cp.plscon.L.Lock()
	defer func() {
		cp.plscon.L.Unlock()
		cp.plscon.Signal()
	}()
	node := cp.pls.Pop()
	if node != nil {
		conn := node.(*client.PoolLinkNode).Conn
		conn.WithTimeout(time.Now().Add(time.Duration(cp.config.MaxIdleConnTimeout)).UnixMilli())
		return conn, nil
	} else {
		conn := reflect.New(cp.connType.Elem()).Interface().(client.IConnect)
		err := conn.Dial(cp.address, cp.config.MaxIdleConnTimeout)
		if err != nil {
			return nil, err
		}
		node = &client.PoolLinkNode{Conn: conn}
		conn.WithAffiliation(cp)
		conn.WithTimeout(time.Now().Add(time.Duration(cp.config.MaxIdleConnTimeout)).UnixMilli())
		conn.WithNode(node)
		cp.openingConns++
		cp.pls.Push(node)
		return conn, nil
	}
}

func (cp *ConnectPool) Put(conn client.IConnect) error {
	cp.plscon.L.Lock()
	defer func() {
		cp.plscon.L.Unlock()
		cp.plscon.Signal()
	}()
	// 超过最大空闲连接数，关闭连接
	if cp.openingConns >= cp.maxIdleConn {
		cp.Close(conn)
		cp.Remove(conn.Node())
		return nil
	}
	cp.pls.Push(&client.PoolLinkNode{Conn: conn})
	return nil
}

func (cp *ConnectPool) Len() int {
	return cp.openingConns
}

func (cp *ConnectPool)Remove(node intrusive.INode){
	cp.pls.Remove(node)
	cp.openingConns--
}

func (cp *ConnectPool) Close(conn client.IConnect) error {
	if conn == nil {
		return errors.New("connection is nil")
	}
	conn.Close()
	return nil
}