package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/rpc/client/msn"
	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

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
	}

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
	return atomic.LoadInt32(&c.state) == LCS_Connected
}

func (c *LongConn) RequestMessage(message proto.Message, timeout int64) (proto.Message, error) {
	if c.currentGoroutineId == utils.GetCurrentGoroutineID() {
		vlog.Error(fmt.Sprintf("RequestMessage cannot block calls in its own thread %v",  message))
		return nil ,fmt.Errorf("RequestMessage cannot block calls in its own thread %v",  message)
	}

	if !c.IsConnected() {
		return nil, errors.New("connect closed")
	}

	msgAny, err := anypb.New(message)
	if err != nil {
		vlog.Error(fmt.Sprintf("anypb.New err %v  msg %v",  err,message))
		return nil ,fmt.Errorf("anypb.New err %v  msg %v",  err,message)
	}

	if msgAny == nil{
		vlog.Error(fmt.Sprintf("RequestMessage message %v is nil",message))
		return nil ,fmt.Errorf("RequestMessage message %v is nil",message)
	}

	seq := msn.Instance().NextId()
	req := &messages.RpcRequestMessage{
		SequenceID:  seq,
		ForwardTime: uint64(time.Now().UnixMilli()),
		Timeout:     uint64(timeout),
		Message:     msgAny,
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

	b, err := messages.MarshalRequestProtobuf(req.SequenceID, req.Timeout, req.Message)
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

	var result proto.Message
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

	readbuffer := circbuf.NewLinkBuffer(4096)

	defer func() {
		defer readbuffer.Close()
		c.done.Done()
	}()

	var (
		nRead          int
		readTempBuffer [1024]byte
		byteKleepAlive []byte
		err            error

		offsetWrite int
		nWrite      int

		msg interface{}
	)
	for {

		nRead, err = c.conn.Read(readTempBuffer[:])
		if err != nil {
			goto exit_reader_lable
		}

		offsetWrite = 0
		for {
			nWrite, _ = readbuffer.WriteBinary(readTempBuffer[offsetWrite:nRead])
			offsetWrite += nWrite
			readbuffer.Flush()

			_, msg, err = messages.UnMarshalProtobuf(readbuffer)
			if err != nil {
				vlog.Debugf("rpc-long-conn UnMarshal Proto fail error:%s", err.Error())
				goto exit_reader_lable
			}

			if nWrite == 0 && msg == nil {
				goto exit_reader_lable
			}

			switch pingMessage := msg.(type) {
			case *messages.RpcPingMessage:
				pingMessage.VerifyKey += 1
				byteKleepAlive, err =messages.MarshalPingProtobuf(pingMessage.VerifyKey+1)
				if err != nil {
					vlog.Debugf("rpc-long-conn Marshal KleepAlive Message fail error:%s", err)
					goto exit_reader_lable
				}
				
				_, err = c.conn.Write(byteKleepAlive)
				if err != nil {
					vlog.Debugf("rpc-long-conn Response KleepAlive Message fail error:%s", err)
					goto exit_reader_lable
				}
			default:
				if msg != nil {
					c.mailbox <- msg
				}
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

	if msg.Result != nil {
		result, err := msg.Result.UnmarshalNew()
		if err != nil {
			future.cond.L.Unlock()
			panic(err)
		}

		future.result = result
		switch resultType := result.(type) {
		case *messages.RpcError:
			future.err = errors.New(resultType.Err)
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
