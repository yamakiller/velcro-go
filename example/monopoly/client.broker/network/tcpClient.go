package network

import (
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"time"

	"github.com/yamakiller/magicLibs/actors"
	"github.com/yamakiller/magicLibs/boxs"
	"github.com/yamakiller/magicNet/netmsgs"
	protomessge "github.com/yamakiller/velcro-go/cluster/gateway/protomessage"
	"github.com/yamakiller/velcro-go/example/monopoly/client.broker/player"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"google.golang.org/protobuf/proto"
)

func NewTCPClient() (*TCPClient,error) {
	client := &TCPClient{
		Box: boxs.SpawnBox(nil),
		core: actors.New(nil),
		_queue: make(chan interface{},256),
		_closed: make(chan bool),
	}
	client.Box.Register(reflect.TypeOf(&netmsgs.Message{}), client.onMessage)
	_, err := client.core.New(func(pid *actors.PID) actors.Actor{
		client.Box.WithPID(pid)
		return client.Box
	},0)
	if err != nil {
		return nil, err
	}
	player.StartUp(client.Core())
	
	return client,nil
}

// TCPClient
type TCPClient struct {
	Addr string
	*boxs.Box
	core    *actors.Core
	conn    net.Conn
	_queue   chan interface{}
	recvice        *circbuf.LinkBuffer
	_closed  chan bool
	_state   int
	secret []byte
}

func (slf *TCPClient) Core()*actors.Core {
	return slf.core
}

func (slf *TCPClient) Connect(addr string, timeout time.Duration) error {
	if slf._closed != nil {
		slf._closed = make(chan bool, 1)
	}
	c, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}

	slf.conn = c

	go slf.reader()
	go slf.sender()

	return err
}
func (slf * TCPClient) reader(){
	offset := 0
	var tmp [1024]byte
	for {
		if slf.isStoped(){
			return
		}
		var (
			n    int   = 0
			werr error = nil
		)
		rl,err := slf.conn.Read(tmp[:])
		if err != nil {
			continue
		}
		n, werr = slf.recvice.WriteBinary(tmp[:rl])
		offset += n
		if err := slf.recvice.Flush(); err != nil {
			return
		}
		
		msg, err := slf.UnSeria()
		if msg != nil{
			slf.GetPID().Post(&netmsgs.Message{Sock: slf.Addr,
				Data: msg})
		}

		if msg == nil {
			if werr != nil {
				slf._closed <- true
				return
			}
			continue
		}
		if err != nil {
			if err == io.EOF {
				continue
			}
			slf._closed <- true
			break
		}
	}
}
func (slf *TCPClient) sender(){
	defer func() {
		slf._state = 1
		slf.conn.Close()
	}()

	for {
		select {
		case <-slf._closed:
			goto exit
		case msg := <-slf._queue:
			if err := slf.Seria(msg); err != nil {
				fmt.Println("write error,", err)
				goto exit
			}
		}
	}
	exit:
	slf._closed <- true
}
func (slf *TCPClient) isStoped()bool{
	select{
	case <-slf._closed:
			return true
	default:
		return false
	}
}
func (slf *TCPClient) Wait() {
	for {
		select {
		case <-slf._closed:
			// player.Instance().RemovePlayer(slf.Addr)
			return
		}
	}
}
const (
	HeaderSize = 2
)
func (slf *TCPClient) UnSeria() (interface{}, error) {
	return protomessge.UnMarshal(slf.recvice,slf.secret)
}

func (slf *TCPClient) Seria(msg interface{}) error {
	d,e := protomessge.Marshal(msg.(proto.Message),slf.secret)
	if e != nil{
		return e
	}
	if _,e := slf.conn.Write(d); e != nil {
		return e
	}
	return nil
}

func (slf *TCPClient) Push(data interface{}) error {
	if slf._state != 0 {
		return errors.New("connection closed")
	}
	slf._queue <- data
	return nil
}

func (slf *TCPClient) Close() {
	slf._state = 1
	if slf.conn != nil {
		slf.conn.Close()
	}

	if slf._closed != nil {
		slf._closed <- true
	}
}
func (slf *TCPClient) RequestMessage(message interface{}) error {
	if err := slf.Push(message); err != nil {
		return err
	}
	return nil
}

func (slf *TCPClient)onMessage(context *boxs.Context) {
	request := context.Message().(*netmsgs.Message)
	if request == nil {
		context.Error("message error")
		return
	}

	handle := player.Instance().GetPlayer(request.Sock.(string))
	if handle == nil {
		context.Error("socket not found:%d", request.Sock)
		return
	}
	
	// evt := request.Data.(*event.Handle)
	// util.AssertEmpty(evt, "event handle is null")
	// dec := protos.FlatDecoder{}
	// flatData, err := dec.UnMarshal(evt.Desc, evt.Data)
	// if err != nil {
	// 	context.Error("socket decode error: %s ", request.Sock, err.Error(), evt.Desc)
	// 	return
	// }

	// handle.GetPID().Post(flatData)
	// context.Debug("message post %v", flatData)
}