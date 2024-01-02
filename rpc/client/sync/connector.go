package sync

import (
	"errors"
	"net"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/rpc/errs"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/syncx"
	"google.golang.org/protobuf/proto"
)

func NewConn() *Conn {
	return &Conn{
		sequence: 0,
	}
}

type Conn struct {
	conn     net.Conn
	sequence int32
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
func (rc *Conn) RequestMessage(message proto.Message, timeout uint64) (proto.Message, error) {
	seq := rc.nextID()
	forwardTime := uint64(time.Now().UnixMilli())
	b, err := messages.MarshalRequestProtobuf(seq, timeout, message)
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
			diff := int64(timeout) - (time.Now().UnixMilli() - int64(forwardTime))
			rc.conn.SetReadDeadline(time.Now().Add(time.Duration(diff) * time.Millisecond))
		}

		nr, err := rc.conn.Read(readtemp[:])
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				return nil, errs.ErrorRequestTimeout
			}
			return nil, err
		}

		offset := 0
		for {
			nw, err := readbuffer.Write(readtemp[offset:nr])
			offset += nw

			_, msg, uerr := messages.UnMarshalProtobuf(readbuffer)
			if uerr != nil {
				return nil, uerr
			}

			if err != nil && msg == nil {
				// 数据包存在问题
				return nil, err
			}

			if msg != nil {
				switch message := msg.(type) {
				case *messages.RpcPingMessage:
				case *messages.RpcResponseMessage:
					if message.SequenceID == seq {
						result, err := message.Result.UnmarshalNew()
						if err != nil {
							return nil, err
						}
						switch r := result.(type) {
						case *messages.RpcError:
							return nil, errors.New(r.Err)
						default:
							return r, nil
						}
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
