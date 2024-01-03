package clientpool

import (
	"errors"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/yamakiller/velcro-go/utils"
)

type Connect interface{
	Connect(addr string, timeout time.Duration) error
	RequsetMessage(proto.Message)error
	Ping() error
	Close() error
}

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

func NewConnectPool(serviceName string, idlConfig IdleConfig) *ConnectPool{
	res := &ConnectPool{}
	res.pool = NewPool(res.connect ,res.close,idlConfig.MaxIdlePerAddress,idlConfig.MaxMessageTimeout)
	getSharedTicker1(res,idlConfig.MaxIdleTimeout)
	return res
}

type ConnectPool struct {
	pool *Pool
	DialTimeout func(string,time.Duration)(Connect,error)
	address string
}

func (cp *ConnectPool) RequsetMessage(msg proto.Message) error{
	conn,err := cp.pool.Get()
	if err != nil{
		return err
	}
	err = conn.(Connect).RequsetMessage(msg)
	if err == errors.New("closed"){
		cp.pool.Close(conn)
	}else{
		cp.pool.Put(conn)
	}
	return nil
}

func (cp *ConnectPool) connect(timeout time.Duration) (any, error){
	conn,err := cp.DialTimeout(cp.address,timeout)
	if err != nil{
		return nil, err
	}
	return conn, nil
}

func (cp *ConnectPool) close(conn interface{}) error{
	return conn.(Connect).Close()
}

func (cp *ConnectPool) Tick(){
	len := cp.pool.Len()
	for i:=0;i<len;i++{
		conn,err :=cp.pool.Get()
		if err != nil{
			return
		}

		if conn != nil{
			if err := conn.(Connect).Ping(); err != nil{
				cp.pool.Close(conn)
			}else{
				cp.pool.Put(conn)
			}
		}
	}
}