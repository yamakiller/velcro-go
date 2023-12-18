package rpc

import (
	"net"
	"sync"
	"time"

	"github.com/yamakiller/velcro-go/cluster/rpc/rpcmessage"
	"github.com/yamakiller/velcro-go/containers"
)

type RpcClient struct {
	conn    net.Conn
	sendbox *containers.Queue
	sendcon *sync.Cond
	stopper chan struct{}
	done    sync.WaitGroup

	MarshalRequest rpcmessage.MarshalRequestFunc
	MarshalMessage rpcmessage.MarshalMessageFunc
	MarshalPing    rpcmessage.MarshalPingFunc
}

func (rc *RpcClient) Dial(addr string, timeout time.Duration) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	rc.conn, err = net.DialTimeout("tcp", address.AddrPort().String(), timeout)
	if err != nil {
		return err
	}

	//1.启动发送及接收
	rc.done.Add(2)
	go func() {
		defer rc.done.Done()
		for {
			rc.sendcon.L.Lock()
			if !rc.isStopped() {
				rc.sendcon.Wait()
			}
			rc.sendcon.L.Unlock()

			for {
				msg, ok := rc.sendbox.Pop()
				if ok && msg == nil {
					goto exit_lable
				}

				if !ok {
					break
				}

				var b []byte
				switch message := msg.(type) {
				case *rpcmessage.RpcRequestMessage:
					b, err = rc.MarshalRequest(message.SequenceID, message.Timeout, message.Message)
					break
				case *rpcmessage.RpcMsgMessage:
					b, err = rc.MarshalMessage(message.SequenceID, message.Message)
					break
				case *rpcmessage.RpcPingMessage:
					b, err = rc.MarshalPing(message.VerifyKey)
					break
				default:
					panic("unknown rpc message")
				}

				_, err := rc.conn.Write(b)
				if err != nil {
					goto exit_lable
				}
			}
		}
	exit_lable:
	}()

	return nil
}

func (rc *RpcClient) isStopped() bool {
	select {
	case <-rc.stopper:
		return false
	default:
		return true
	}
}
