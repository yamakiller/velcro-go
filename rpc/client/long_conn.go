package client

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/rpc/client/msn"
	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
	"github.com/yamakiller/velcro-go/vlog"

	cmap "github.com/orcaman/concurrent-map"
)

type LongConnState int

const (
	LCS_Disconnected = iota
	LCS_Connecting
	LCS_Connected
	LCS_Disconnecting
)

type LongConnDirect int

const (
	LCS_Unknown = iota
	LCS_Idle
	LCS_Busy
)

func NewLongConn(ascription LongConnPool, usedLastTime int64) *LongConn {
	conn := &LongConn{
		ascription:   ascription,
		mailbox:      make(chan interface{}, 1),
		reqsbox:      cmap.New(),
		state:        LCS_Disconnected,
		direct:       LCS_Unknown,
		usedLastTime: usedLastTime,
		request:      protocol.NewBinaryProtocol(),
		response:     protocol.NewBinaryProtocol(),
	}
	conn.processor = messages.NewRpcServiceProcessor(conn)
	return conn
}

type LongConnLinkedNode struct {
	intrusive.LinkedNode
}

type LongConn struct {
	LongConnLinkedNode
	ascription   LongConnPool
	conn         net.Conn
	done         sync.WaitGroup
	mailbox      chan interface{}
	guardianDone sync.WaitGroup
	reqsbox      cmap.ConcurrentMap

	currentGoroutineId int
	state              int32
	direct             int32
	usedLastTime       int64 // 最后使用时间,可以利用此计算连接闲置时间
	request            *protocol.BinaryProtocol
	response           *protocol.BinaryProtocol
	processor          thrift.TProcessor
}

func (c *LongConn) Dial(addr string, timeout time.Duration) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	c.state = LCS_Connecting
	c.conn, err = net.DialTimeout("tcp", address.AddrPort().String(), timeout)
	if err != nil {
		c.state = LCS_Disconnected
		return err
	}

	c.done.Add(1)
	c.guardianDone.Add(1)

	gofunc.RecoverGoFuncWithInfo(context.Background(), c.reader,
		gofunc.NewBasicInfo("rpc-long-conn-reader", c.EscalateFailure))

	gofunc.RecoverGoFuncWithInfo(context.Background(), c.guardian,
		gofunc.NewBasicInfo("rpc-long-conn-guardian", c.EscalateFailure))

	c.state = LCS_Connected
	return nil
}

func (c *LongConn) IsConnected() bool {
	if atomic.LoadInt32(&c.state) == LCS_Connected {
		return true
	}
	return false
}

func (c *LongConn) RequestMessage(message thrift.TStruct, name string, timeout int64) ([]byte, error) {
	if c.currentGoroutineId == utils.GetCurrentGoroutineID() {
		panic("RequestMessage cannot block calls in its own thread")
	}

	if !c.IsConnected() {
		return nil, errors.New("connect closed")
	}
	seq := msn.Instance().NextId()
	msg,err :=  messages.MarshalTStruct(context.Background(),c.request,message,name,seq)
	if err != nil{
		return nil,err
	}

	req := &messages.RpcRequestMessage{
		SequenceID:  seq,
		ForwardTime: int64(time.Now().UnixMilli()),
		Timeout:     int64(timeout),
		Message:     msg,
	}

	msg,err =  messages.MarshalTStruct(context.Background(),c.request,req,protocol.MessageName(req),seq)
	if err != nil{
		return nil,err
	}

	future := &Future{
		sequenceID: seq,
		cond:       sync.NewCond(&sync.Mutex{}),
		done:       false,
		request:    req,
		result:     nil,
		err:        nil,
		t:          time.NewTimer(time.Duration(timeout) * time.Millisecond),
	}
	
	b, err := messages.Marshal(msg)
	if err != nil {
		future.cond = nil
		future.request = nil
		future.t.Stop()
		return nil, err
	}

	c.reqsbox.SetIfAbsent(strconv.FormatInt(int64(future.sequenceID), 10), future)
	if timeout > 0 {
		tp := time.AfterFunc(time.Duration(timeout)*time.Millisecond, func() {
			future.cond.L.Lock()
			if future.done {
				future.cond.L.Unlock()
				return
			}
			future.err = errs.ErrorRequestTimeout
			future.done = true

			c.reqsbox.Remove(strconv.FormatInt(int64(future.sequenceID), 10))
			future.cond.L.Unlock()
			future.cond.Signal()
		})
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&future.t)), unsafe.Pointer(tp))
	}

	gofunc.GoFunc(context.Background(), func() {
		_, cErr := c.conn.Write(b)
		if cErr != nil {
			future.cond.L.Lock()
			if !future.done {
				future.done = true
				future.err = cErr
				tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&future.t))))
				if tp != nil {
					tp.Stop()
				}
				c.reqsbox.Remove(strconv.FormatInt(int64(future.sequenceID), 10))
				future.cond.Signal()
			}
			future.cond.L.Unlock()
		}
	})

	var result []byte
	var resultErr error
	if timeout > 0 {
		future.cond.L.Lock()
		if !future.done {
			future.cond.Wait()
		}

		result = future.result
		resultErr = future.err

		future.cond.L.Unlock()
	}

	return result, resultErr
}

func (c *LongConn) reader() {
	recvice := circbuf.NewLinkBuffer(4096)
	readbuffer := protocol.NewBinaryProtocol()
	defer func() {
		defer readbuffer.Close()
		c.done.Done()
	}()

	var (
		nRead          int
		readTempBuffer [1024]byte
		err            error

		offsetWrite int
		nWrite      int
	)
	for {

		nRead, err = c.conn.Read(readTempBuffer[:])
		if err != nil {
			goto exit_reader_lable
		}

		offsetWrite = 0
		for {
			nWrite, err = recvice.WriteBinary(readTempBuffer[offsetWrite:])
			if err != nil{
				goto exit_reader_lable
			}
			offsetWrite += nWrite
			if err := recvice.Flush(); err != nil {
				goto exit_reader_lable
			}
			msg,err :=  messages.UnMarshal(recvice)
			if msg == nil || err != nil{
				goto exit_reader_lable
			}
			readbuffer.Release()
			readbuffer.Write(msg)
			name, _, seq, err := readbuffer.ReadMessageBegin(context.Background())
			if err != nil {
				vlog.Debugf("rpc-long-conn UnMarshal Proto fail error:%s", err.Error())
				goto exit_reader_lable
			}
			switch name {
			case "messages.RpcPingMessage":
				ping := messages.NewRpcPingMessage()
				ping.Read(context.Background(), readbuffer)
				ping.VerifyKey += 1
				m ,err := messages.MarshalTStruct(context.Background(),readbuffer, ping,protocol.MessageName(ping), seq)
				if err !=nil{
					goto exit_reader_lable
				}
				b,err := messages.Marshal(m)
				if err != nil{
					goto exit_reader_lable
				}
				_, err = c.conn.Write(b)
				if err != nil {
					vlog.Debugf("rpc-long-conn Response KleepAlive Message fail error:%s", err)
					goto exit_reader_lable
				}
				
			case "messages.RpcResponseMessage":
				response := messages.NewRpcResponseMessage()
				if err := response.Read(context.Background(), readbuffer); err == nil {
					c.mailbox <- response
				}
			default:
			}
			if offsetWrite == nRead {
				break
			}
		}
	}
exit_reader_lable:
	atomic.StoreInt32(&c.state, LCS_Disconnecting)
	c.conn.Close()
	c.mailbox <- nil
}

func (c *LongConn) guardian() {
	c.currentGoroutineId = utils.GetCurrentGoroutineID()
	defer func() {
		c.guardianDone.Done()
	}()
	for {
		msg, ok := <-c.mailbox
		if !ok {
			goto exit_guardian_lable
		}

		if msg == nil {
			break
		}

		switch message := msg.(type) {
		case *messages.RpcResponseMessage:
			c.onResponse(message)
		default:
		}
	}
exit_guardian_lable:
	if c.ascription != nil {
		c.ascription.Discard(c)
	}

	c.done.Wait() // 等待读写线程结束
	close(c.mailbox)

	maps := c.reqsbox.Items()
	c.reqsbox.Clear()
	for _, v := range maps {
		future := v.(*Future)
		future.cond.L.Lock()
		if !future.done {
			future.done = true
			future.err = errors.New("connect closed")
			future.result = nil
			future.cond.Signal()
		}
		future.cond.L.Unlock()
	}
	atomic.StoreInt32(&c.state, LCS_Disconnected)
}

func (c *LongConn) onResponse(msg *messages.RpcResponseMessage) {
	future := c.getFuture(msg.SequenceID)
	if future == nil {
		return
	}

	future.cond.L.Lock()
	if future.done {
		future.cond.L.Unlock()
		return
	}

	if msg.Result_ != nil {
		c.response.Release()
		c.response.Write(msg.Result_)
		name, _, _, err := c.response.ReadMessageBegin(context.Background())
		if err != nil {
			future.cond.L.Unlock()
			panic(err)
		}
		future.result = msg.Result_
		switch name {
		case "messages.RpcError":
			{
				c.response.Release()
				rpcerr := messages.NewRpcError()
				rpcerr.Write(context.Background(), c.response)
				future.err = errors.New(rpcerr.Err)
			}
		default:
			future.err = nil
		}
	}

	future.done = true

	tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&future.t))))
	if tp != nil {
		tp.Stop()
	}
	future.cond.Signal()
	future.cond.L.Unlock()
}

func (c *LongConn) EscalateFailure(reason interface{}, message interface{}) {
	vlog.Errorf("%s \nstack%s", reason.(error).Error(), message.(string))
}

func (c *LongConn) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	c.done.Wait()
}

func (c *LongConn) getFuture(id int32) *Future {
	result, ok := c.reqsbox.Pop(strconv.FormatInt(int64(id), 10))
	if !ok {
		return nil
	}
	return result.(*Future)
}
