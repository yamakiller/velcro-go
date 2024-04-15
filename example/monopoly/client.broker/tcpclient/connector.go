package tcpclient

import (
	"context"
	"fmt"
	"strconv"

	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/apache/thrift/lib/go/thrift"
	cmap "github.com/orcaman/concurrent-map"
	protomessge "github.com/yamakiller/velcro-go/cluster/gateway/protomessage"
	"github.com/yamakiller/velcro-go/cluster/protocols/pubs"
	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/rpc/client/msn"
	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/collection/intrusive"
	"github.com/yamakiller/velcro-go/vlog"
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
		proto: protocol.NewBinaryProtocol(),
		mailbox: make(chan interface{}, 1),
		reqsbox: cmap.New(),
		state:   LCS_Disconnected,
		methods: make(map[interface{}]thrift.TStruct),
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
	proto        *protocol.BinaryProtocol
	currentGoroutineId int
	state              int32
	secret             []byte
	sequenceID         int32

	methods map[interface{}]thrift.TStruct
}

func (c *Conn) Dial(addr string, timeout time.Duration) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	c.state = LCS_Connecting
	c.conn, err = net.DialTimeout("tcp", address.AddrPort().String(), timeout)
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

func (c *Conn) Register(key string, value thrift.TStruct) {
	c.methods[key] = value
}

func (c *Conn) IsConnected() bool {
	return atomic.LoadInt32(&c.state) == LCS_Connected
}

func (c *Conn) RequestMessage(message []byte, timeout int64) (thrift.TStruct, error) {
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

	b, err := protomessge.Marshal(message, c.secret)
	if err != nil {
		future.cond = nil
		future.request = nil
		future.t.Stop()
		return nil, err
	}
	c.proto.Release()

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

	var result thrift.TStruct
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
			c.proto.Write(msg)
			name,_,seqId,err := c.proto.ReadMessageBegin(context.Background())
			switch name {
			case "pubs.PingMsg":
				res := pubs.NewPingMsg()
				res.Read(context.Background(),c.proto)
				c.proto.Release()
				res.VerificationKey = res.VerificationKey+1
				b,err := messages.MarshalTStruct(context.Background(),c.proto,res,protocol.MessageName(res),seqId)
				if err != nil{
					goto exit_reader_lable
				}
				byteKleepAlive, err = protomessge.Marshal(b, c.secret)
				if err != nil {
					goto exit_reader_lable
				}
				_, err = c.conn.Write(byteKleepAlive)
				if err != nil {
					vlog.Debugf("rpc-long-conn Response KleepAlive Message fail error:%s", err)
					goto exit_reader_lable
				}
			default:
				if m,ok :=c.methods[name];ok{
					m.Read(context.Background(),c.proto)
					c.proto.Release()
					c.mailbox <- m
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
		case *pubs.PubkeyMsg:
			c.onPubkeyReply(message)
		case *pubs.Error:
			c.onError(message)
		default:
			c.onResponse(message.(thrift.TStruct))
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
func (c *Conn) onError(err *pubs.Error) {
	future := c.getFuture(c.sequenceID)
	if future == nil {
		return
	}
	future.cond.L.Lock()
	if future.done {
		future.cond.L.Unlock()
		return
	}
	if err != nil {
		future.result = nil

		future.err = fmt.Errorf("message %v error:%v", err.Name, err.Err)
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

func (c *Conn) onResponse(msg thrift.TStruct) {
	future := c.getFuture(c.sequenceID)
	if future == nil {
		return
	}

	// if m,ok := c.methods[reflect.TypeOf(msg)]; ok {
	// 	go m.value(context.Background(),msg)
	// 	return
	// }

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
func (dl *Conn) onPubkeyReply(message *pubs.PubkeyMsg) {

	// var (
	// 	prvKey crypto.PrivateKey
	// 	pubKey crypto.PublicKey
	// )

	// pubkeyByte, err := base64.StdEncoding.DecodeString(message.Key)
	// if err != nil {
	// 	vlog.Debugf("public key decode error %s", err.Error())
	// 	// ctx.Close(ctx.Self())
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
