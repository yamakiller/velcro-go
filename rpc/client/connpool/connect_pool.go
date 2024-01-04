package clientpool

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
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
	New     func(string, time.Duration) (IConnect, error)
	address string
	config  IdleConfig
}

func (cp *ConnectPool) RequsetMessage(msg proto.Message) (interface{}, error) {
	var (
		conn IConnect
		res interface{}
		err error
	)
	conn, err = cp.Get()
	if err != nil {
		return nil, err
	}
	res, err = conn.RequsetMessage(msg,cp.config.MaxMessageTimeout.Milliseconds())
	if err == errors.New("closed") {
		cp.Close(conn)
	} else {
		cp.Put(conn)
	}
	return res,nil
}

func (cp *ConnectPool) Close(conn interface{}) error {
	if conn == nil {
		return errors.New("connection is nil")
	}
	conn.(IConnect).Close()
	cp.openingConns--
	return nil
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
		if node.(IConnect).Timeout()-time.Now().UnixMilli() <= 0 {
			remove_list = append(remove_list, node)
		}
		return true
	})

	for _, node := range remove_list {
		cp.Close(node)
		cp.pls.Remove(node)
	}
}


func (cp *ConnectPool) Get() (IConnect, error) {
	cp.plscon.L.Lock()
	defer func() {
		cp.plscon.L.Unlock()
		cp.plscon.Signal()
	}()
	node := cp.pls.Pop()
	if node != nil {
		conn := node.(*PoolLinkNode).value
		conn.WithTimeout(time.Now().Add(time.Duration(cp.config.MaxIdleConnTimeout)).UnixMilli())
		return conn, nil
	} else {
		conn := reflect.New(cp.connType.Elem()).Interface().(IConnect)
		err := conn.Dial(cp.address, cp.config.MaxIdleConnTimeout)
		if err != nil {
			return nil, err
		}
		node = &PoolLinkNode{value: conn}
		conn.WithAffiliation(cp)
		conn.WithTimeout(time.Now().Add(time.Duration(cp.config.MaxIdleConnTimeout)).UnixMilli())
		conn.WithNode(node)
		cp.openingConns++
		cp.pls.Push(node)
		return conn, nil
	}
}

func (cp *ConnectPool) Put(conn IConnect) error {
	cp.plscon.L.Lock()
	defer func() {
		cp.plscon.L.Unlock()
		cp.plscon.Signal()
	}()
	// 超过最大空闲连接数，关闭连接
	if cp.openingConns >= cp.maxIdleConn {
		cp.Close(conn)
		return nil
	}
	cp.pls.Push(&PoolLinkNode{value: conn})
	return nil
}

func (cp *ConnectPool) Len() int {
	return cp.openingConns
}