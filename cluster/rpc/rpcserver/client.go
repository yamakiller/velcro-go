package rpcserver

import (
	"context"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/cluster/rpc/rpcmessage"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/circbuf"
)

type RpcClient struct {
	parent   *RpcServer
	clientID *network.ClientID   // 客户端ID
	recvice  *circbuf.RingBuffer // 接收缓冲区
	requests map[interface{}]func(ctxtimeout context.Context,
		ctx network.Context,
		message interface{}) interface{}
	messages map[interface{}]func(ctx network.Context, message interface{})

	reference int32 //引用计数器
}

func (rc *RpcClient) Accept(ctx network.Context) {
	rc.clientID = ctx.Self()
	rc.reference = 1

	rc.parent.Register(ctx.Self(), rc)
}

func (rc *RpcClient) Ping(ctx network.Context) {
	// 不主动处理心跳
}

func (rc *RpcClient) Recvice(ctx network.Context) {

	offset := 0
	for {
		n, err := rc.recvice.Write(ctx.Message()[offset:])
		offset += n

		_, msg, msgErr := rc.parent.UnMarshal(rc.recvice)
		if msgErr != nil {
			ctx.Close(ctx.Self())
			return
		}

		if msg == nil && err != nil {
			// 属于缓冲区溢出
			ctx.Close(ctx.Self())
			return
		}

		if offset == len(ctx.Message()) {
			break
		}

		if msg == nil {
			continue
		}

		switch message := msg.(type) {
		case *rpcmessage.RpcPingMessage:
			rc.onRpcPing(ctx, message)
		case *rpcmessage.RpcRequestMessage:
			f, ok := rc.requests[reflect.TypeOf(msg)]
			if !ok {
				continue
			}

			timeout := int64(message.Timeout) - time.Now().UnixMilli() - int64(message.ForwardTime)

			// 如果已超时
			if timeout <= 0 {
				//rc.onTimeout(ctx, message.SequenceID)
				continue
			}
			// 设置超时器
			ctxout, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
			result := f(ctxout, ctx, msg)
			// 判断是否超时
			select {
			case <-ctxout.Done():
				//rc.onTimeout(ctx, message.SequenceID)
			default:
				if result != nil {
					b, _ := rc.parent.MarshalResponse(message.SequenceID, 0, result)
					if b != nil {
						ctx.PostMessage(ctx.Self(), b)
					}
				}
			}
			cancel()
		case *rpcmessage.RpcMsgMessage:
			f, ok := rc.messages[reflect.TypeOf(msg)]
			if !ok {
				continue
			}
			f(ctx, msg)
		default:
			ctx.Debug("unknown RPC message")
		}

	}
}

func (rc *RpcClient) Closed(ctx network.Context) {
	rc.parent.UnRegister(ctx.Self())
}

func (rc *RpcClient) onRpcPing(ctx network.Context, message *rpcmessage.RpcPingMessage) {

	b, err := rc.parent.MarshalPing(message.VerifyKey + 1)
	if err != nil {
		ctx.Close(ctx.Self())
		return
	}

	ctx.PostMessage(ctx.Self(), b)
}

/*func (rc *RpcClient) onTimeout(ctx network.Context, sequenceID int32) {
	b, _ := rc.parent.MarshalResponse(sequenceID, -1, nil)
	if b != nil {
		ctx.PostMessage(ctx.Self(), b)
	}
}*/

// RefInc 引用计数器+1
func (rc *RpcClient) referenceIncrement() int32 {
	return atomic.AddInt32(&rc.reference, 1)
}

// RefDec 引用计数器-1
func (rc *RpcClient) referenceDecrement() int32 {
	return atomic.AddInt32(&rc.reference, -1)
}
