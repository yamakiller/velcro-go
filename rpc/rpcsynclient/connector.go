package rpcsynclient

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/rpc"
	"github.com/yamakiller/velcro-go/rpc/rpcmessage"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/syncx"
)

func NewConn(options ...ConnConfigOption) *Conn {
	config := Configure(options...)

	return NewConnConfig(config)
}

func NewConnConfig(config *ConnConfig) *Conn {
	return &Conn{
		sequence:       0,
		MarshalRequest: config.MarshalRequest,
		UnMarshal:      config.UnMarshal,
	}
}

type Conn struct {
	conn     net.Conn
	sequence int32

	MarshalRequest rpcmessage.MarshalRequestFunc
	UnMarshal      rpcmessage.UnMarshalFunc
}

func (rc *Conn) Dial(addr string, timeout time.Duration) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	rc.conn, err = net.DialTimeout("tcp", address.AddrPort().String(), timeout)
	if err != nil {
		return err
	}

	return nil
}

// RequestMessage 请求消息并等待回复，超时时间单位为毫秒
func (rc *Conn) RequestMessage(message interface{}, timeout uint64) (interface{}, error) {
	seq := rc.nextID()
	req := &rpcmessage.RpcRequestMessage{
		SequenceID:  seq,
		ForwardTime: uint64(time.Now().UnixMilli()),
		Timeout:     timeout,
		Message:     message,
	}

	b, err := rc.MarshalRequest(req.SequenceID, req.Timeout, req.Message)
	if err != nil {
		return nil, err
	}

	_, err = rc.conn.Write(b)
	if err != nil {
		return nil, err
	}

	var readtemp [1024]byte
	readbuffer := circbuf.New(32768, &syncx.NoMutex{})

	for {

		if timeout > 0 {
			diff := int64(req.Timeout) - (time.Now().UnixMilli() - int64(req.ForwardTime))
			rc.conn.SetReadDeadline(time.Now().Add(time.Duration(diff) * time.Millisecond))
		}

		nr, err := rc.conn.Read(readtemp[:])
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				return nil, rpc.ErrorRequestTimeout
			}
			return nil, err
		}

		offset := 0
		for {
			nw, err := readbuffer.Write(readtemp[offset:nr])
			offset += nw

			_, msg, uerr := rc.UnMarshal(readbuffer)
			if uerr != nil {
				return nil, uerr
			}

			if err != nil && msg == nil {
				// 数据包存在问题
				return nil, err
			}

			if msg != nil {
				switch message := msg.(type) {
				case *rpcmessage.RpcPingMessage:
				case *rpcmessage.RpcResponseMessage:
					if message.SequenceID == seq {
						if message.Result == -1 {
							return nil, rpc.ErrorRequestTimeout
						}

						return message.Message, nil
					}
				default:
				}
			}

			if offset == nr {
				break
			}
		}
	}
}

func (rc *Conn) Close() {
	if rc.conn != nil {
		rc.conn.Close()
		rc.conn = nil
	}
}

func (rc *Conn) nextID() int32 {
	for {
		if id := atomic.AddInt32(&rc.sequence, 1); id > 0 {
			return id
		} else if atomic.CompareAndSwapInt32(&rc.sequence, id, 1) {
			return 1
		}
	}
}
