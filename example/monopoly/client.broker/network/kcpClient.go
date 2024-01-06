package network

// import (
// 	"bufio"
// 	"errors"
// 	"fmt"
// 	"golang-test/base/protos"
// 	"io"
// 	"net"
// 	"reflect"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/yamakiller/magicLibs/actors"
// 	"github.com/yamakiller/magicLibs/boxs"
// 	"github.com/yamakiller/magicLibs/net/connection"
// 	"github.com/yamakiller/magicLibs/util"
// 	"github.com/yamakiller/magicNet/netmsgs"
// 	"google.golang.org/protobuf/proto"
// 	"google.golang.org/protobuf/reflect/protoreflect"
// 	"modernc.org/libc/netdb"
// )

// func NewKCPClient(box boxs.Box,core *actors.Core) (*KCPClient,error) {
// 	client := &KCPClient{
// 		_core: core,
// 		_queue: make(chan interface{},256),
// 		_closed: make(chan bool),
// 	}
// 	client.Box = box
// 	_, err := client._core.New(func(pid *actors.PID) actors.Actor{
// 		client.Box.WithPID(pid)
// 		return &client.Box
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	client._conn = &connection.KCPClient{}
// 	client._conn.S = &client
// 	client._conn.E = &client
// 	return client,nil
// }

// // KCPClient
// type KCPClient struct {
// 	Addr string
// 	boxs.Box
// 	_core    *actors.Core
// 	_conn connection.KCPClient

// 	// _conn    io.ReadWriteCloser
// 	// _queue   chan interface{}
// 	// _reader  *bufio.Reader
// 	// _writer  *bufio.Writer
// 	// _closed  chan bool
// 	// _timeout time.Time
// 	// _check   int
// 	// _state   int
// 	// _wg      sync.WaitGroup
// }

// func (slf *KCPClient) Core()*actors.Core {
// 	return slf._core
// }

// // func (slf *KCPClient) Connect(addr string, timeout time.Duration) error {
// // 	if slf._closed != nil {
// // 		slf._closed = make(chan bool, 1)
// // 	}
// // 	c, err := net.Conn()  //connection.KCPSeria
// // 	if err != nil {
// // 		return err
// // 	}

// // 	slf._conn = c
// // 	slf._reader = bufio.NewReaderSize(slf._conn, 8192)
// // 	slf._writer = bufio.NewWriterSize(slf._conn, 8192)

// // 	slf._wg.Add(1)
// // 	go func() {
// // 		for {
// // 			msg, err := slf.UnSeria()
// // 			if msg != nil{
// // 				slf.GetPID().Post(&netmsgs.Message{Sock: slf.Addr,
// // 					Data: msg})
// // 			}
// // 			if err != nil {
// // 				if err == io.EOF {
// // 					continue
// // 				}
// // 				slf._closed <- true
// // 				break
// // 			}
// // 		}
// // 	}()

// // 	go func() {
// // 		defer func() {
// // 			slf._state = 1
// // 			slf._conn.Close()
// // 			slf._wg.Done()
// // 		}()

// // 		for {
// // 			select {
// // 			case <-slf._closed:
// // 				goto exit
// // 			case msg := <-slf._queue:
// // 				if err := slf.Seria(msg); err != nil {
// // 					fmt.Println("write error,", err)
// // 					goto exit
// // 				}
// // 			}
// // 		}
// // 	exit:
// // 	}()

// // 	return err
// // }
// // func (slf *KCPClient) Wait() {
// // 	for {
// // 		select {
// // 		case <-slf._closed:
// // 			// player.Instance().RemovePlayer(slf.Addr)
// // 			return
// // 		}
// // 	}
// // }
// func (slf *KCPClient) UnSeria() (interface{}, error) {
// 	timespace, flatDesc, flatData, err := protos.UnMarshal(slf._reader, 4096)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Handle{
// 		Desc: flatDesc,
// 		Data: flatData,
// 		Tts:  timespace}, nil
// }

// func (slf *KCPClient) Seria(msg interface{}) error {

// 	var err error
// 	nsm := msg.(*Handle)
// 	util.AssertEmpty(nsm, "seria network data error check is network serve message")
// 	if _, err = protos.Marshal(slf._writer, nsm.Tts, nsm.Desc, nsm.Data); err != nil {
// 		return err
// 	}

// 	if len(slf._queue) > 0 {
// 		goto exit
// 	}

// 	if slf._writer.Buffered() > 0 {
// 		if err = slf._writer.Flush(); err != nil {
// 			return err
// 		}
// 	}
// exit:
// 	return nil
// }

// func (slf *KCPClient) Push(data interface{}) error {
// 	if slf._state != 0 {
// 		return errors.New("connection closed")
// 	}
// 	slf._queue <- data
// 	return nil
// }

// func (slf *KCPClient) Close() {
// 	slf._state = 1
// 	if slf._conn != nil {
// 		slf._conn.Close()
// 	}

// 	if slf._closed != nil {
// 		slf._closed <- true
// 	}
// }
// func (slf *KCPClient) SendData(message interface{}) error {
// 	msg :=message.(protoreflect.ProtoMessage)
// 	data, _ := proto.Marshal(msg)
// 	post := Encoding(strings.Replace(reflect.TypeOf(message).String(), "*", "", 1), data)
// 	if err := slf.Push(post); err != nil {
// 		return err
// 	}
// 	return nil
// }

