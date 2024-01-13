package tcpclient

import (
	"context"
	"fmt"
	"strconv"

	// "crypto"
	// "encoding/base64"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/yamakiller/velcro-go/cluster/gateway/protomessage"
	"github.com/yamakiller/velcro-go/cluster/protocols/pubs"
	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/rpc/client/msn"
	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Closed() {

}

type AcceptMessage struct {
}

type PingMessage struct {
}

type RecviceMessage struct {
	Data []byte
	Addr net.Addr
}

type ClosedMessage struct {
}


func NewConn() *Conn {
	return &Conn{
		mailbox:      make(chan interface{}, 1),
		reqsbox:      cmap.New(),
		state:        LCS_Disconnected,
	}
}

type ConnState int

const (
	LCS_Disconnected = iota
	LCS_Connecting
	LCS_Connected
	LCS_Disconnecting
)

type IntervalLinkNode struct {
	intrusive.LinkedNode
	Value interface{}
}

type Conn struct {
	conn         net.Conn
	done         sync.WaitGroup
	mailbox      chan interface{}
	guardianDone sync.WaitGroup
	reqsbox      cmap.ConcurrentMap

	currentGoroutineId int
	state              int32
	secret             []byte
	sequenceID int32
}

func (c *Conn) Dial(addr string, timeout time.Duration) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	c.state = LCS_Connecting
	c.conn, err = net.DialTimeout("tcp", address.AddrPort().String(), timeout)
	// c.conn, err = net.Dial("tcp", addr)
	if err != nil {
		c.state = LCS_Disconnected
		fmt.Println(err.Error())
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

func (c *Conn) EscalateFailure(reason interface{}, message interface{}) {
	vlog.Errorf("%s \nstack%s", reason.(error).Error(), message.(string))
}


// func (c *Conn) Register(key interface{}, f func(ctx context.Context, message proto.Message) (proto.Message, error)) {
// 	c.methods[reflect.TypeOf(key)] = f
// }

func (c *Conn) IsConnected() bool {
	if atomic.LoadInt32(&c.state) == LCS_Connected {
		return true
	}
	return false
}

func (c *Conn) RequestMessage(message proto.Message, timeout int64) (proto.Message, error) {
	if c.currentGoroutineId == utils.GetCurrentGoroutineID() {
		panic("RequestMessage cannot block calls in its own thread")
	}

	if !c.IsConnected() {
		return nil, errors.New("connect closed")
	}

	seq := msn.Instance().NextId()
	c.sequenceID = seq
	future := &Future{
		sequenceID: seq,
		cond:       sync.NewCond(&sync.Mutex{}),
		done:       false,
		request:    message,
		result:     nil,
		err:        nil,
		t:          time.NewTimer(time.Duration(timeout) * time.Millisecond),
	}

	b, err := protomessge.Marshal(message.(proto.Message), c.secret)
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

func (c *Conn) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	c.done.Wait()
	c.guardianDone.Wait()
}

func (c *Conn) Destory() {
	close(c.mailbox)
	// c.sendbox = nil
}


func (c *Conn) onResponse(msg protoreflect.ProtoMessage) {
	future := c.getFuture(c.sequenceID)
	if future == nil {
		return
	}
	future.cond.L.Lock()
	if future.done {
		future.cond.L.Unlock()
		return
	}

	if msg != nil {
		future.result = msg
	}

	future.done = true

	tp := (*time.Timer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&future.t))))
	if tp != nil {
		tp.Stop()
	}
	c.sequenceID = 0
	future.cond.Signal()
	future.cond.L.Unlock()
	
}

func (c *Conn) reader() {
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

			msg, err := protomessge.UnMarshal(readbuffer, c.secret)
			if err != nil {
				vlog.Debugf("rpc-long-conn UnMarshal Proto fail error:%s", err.Error())
				goto exit_reader_lable
			}

			if nWrite == 0 && msg == nil {
				goto exit_reader_lable
			}
			switch message := msg.(type) {
			case *pubs.PingMsg:
				msg := &pubs.PingMsg{VerificationKey: message.VerificationKey + 1}
				byteKleepAlive, err = protomessge.Marshal(msg, c.secret)
				if err != nil {
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

func (c *Conn) guardian() {
	c.currentGoroutineId = utils.GetCurrentGoroutineID()

	for {
		msg, ok := <-c.mailbox
		if !ok {
			goto exit_guardian_lable
		}

		if msg == nil {
			break
		}

		switch message := msg.(type) {
		case *pubs.PubkeyMsg:
			c.onPubkeyReply(message)
		case *pubs.Error:
			vlog.Errorf("message %v error:%v",message.Name,message.Err)
		default:
			c.onResponse(message.(proto.Message))
		}
	}
exit_guardian_lable:

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

func (dl *Conn) onPubkeyReply(message *pubs.PubkeyMsg) {

	// var (
	// 	prvKey crypto.PrivateKey
	// 	pubKey crypto.PublicKey
	// )

	// pubkeyByte, err := base64.StdEncoding.DecodeString(message.Key)
	// if err != nil {
	// 	vlog.Debugf("public key decode error %s", err.Error())
	// 	ctx.Close(ctx.Self())
	// 	return
	// }

	// prvKey, pubKey, err = dl.gateway.encryption.Ecdh.GenerateKey(rand.Reader)
	// if err != nil {
	// 	vlog.Debugf("generate public/private key error %s", err.Error())
	// }

	// remotePubkey, ok := dl.gateway.encryption.Ecdh.Unmarshal(pubkeyByte)
	// if !ok {
	// 	vlog.Debug("Public key parsing exception")
	// 	ctx.Close(ctx.Self())
	// 	return
	// }

	// secret, err := dl.gateway.encryption.Ecdh.GenerateSharedSecret(prvKey, remotePubkey)
	// if err != nil {
	// 	vlog.Debugf("generate shared secret error %s", err.Error())
	// 	ctx.Close(ctx.Self())
	// 	return
	// }

}

func (c *Conn) getFuture(id int32) *Future {
	result, ok := c.reqsbox.Pop(strconv.FormatInt(int64(id), 10))
	if !ok {
		return nil
	}
	return result.(*Future)
}