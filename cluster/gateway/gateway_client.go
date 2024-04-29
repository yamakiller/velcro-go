package gateway

import (
	"context"
	"math"
	"sync/atomic"

	protomessge "github.com/yamakiller/velcro-go/cluster/gateway/protomessage"
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/protocols/pubs"
	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/network"

	"github.com/yamakiller/velcro-go/rpc/client/msn"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/lang/fastrand"
	"github.com/yamakiller/velcro-go/vlog"
)

type Client interface {
	ClientID() *network.ClientID
	Secret() []byte
	Accept(ctx network.Context)
	Ping(ctx network.Context)
	Post(message []byte) error
	Recvice(ctx network.Context)
	Closed(ctx network.Context)
	Destory()

	alterRule(rule int32)

	//onPingReply(ctx network.Context, message *pubs.PingMsg)
	//onPubkeyReply(ctx network.Context, message *pubs.PubkeyMsg)
	//onRequestMessage(ctx network.Context, message proto.Message)

	referenceIncrement() int32
	referenceDecrement() int32
}

type ClientConn struct {
	gateway  *Gateway
	clientID *network.ClientID
	ruleID   int32  //角色ID
	secret   []byte //密钥
	recvice  *circbuf.LinkBuffer
	//message_agent messageagent.IMessageAgent

	ping uint64
	// requestTimeout int64 //最大超时时间 毫秒级
	reference int32 //引用计数器
}

func (dl *ClientConn) ClientID() *network.ClientID {
	return dl.clientID
}

func (dl *ClientConn) Secret() []byte {
	return dl.secret
}

func (dl *ClientConn) Accept(ctx network.Context) {
	dl.clientID = ctx.Self()
	dl.ruleID = router.NONE_RULE_ID
	if err := dl.gateway.Register(dl.clientID, dl); err != nil {
		// 在线满
		ctx.Close(ctx.Self())
		return
	}
}

func (dl *ClientConn) Ping(ctx network.Context) {
	if dl.ruleID <= router.KEYED_RULE_ID {
		return
	}
	iprot := protocol.NewBinaryProtocol()
	defer iprot.Close()
	dl.ping = fastrand.Uint64n(math.MaxUint64)
	msg := &pubs.PingMsg{VerificationKey: int64(dl.ping)}
	buf, err := messages.MarshalTStruct(context.Background(), iprot, msg, protocol.MessageName(msg), msn.Instance().NextId())
	if err != nil {
		return
	}
	msgb, err := protomessge.Marshal(buf, dl.secret)
	if err != nil {
		vlog.Errorf("ping marshal message [error:%v]", err.Error())
		return
	}
	ctx.PostMessage(ctx.Self(), msgb)
}

func (dl *ClientConn) Post(message []byte) error {

	return dl.clientID.PostUserMessage(message)
}

func (dl *ClientConn) Recvice(ctx network.Context) {
	offset := 0
	for {
		var (
			n    int   = 0
			werr error = nil
		)

		if offset < len(ctx.Message()) {
			n, werr = dl.recvice.WriteBinary(ctx.Message()[offset:])
			offset += n
			if err := dl.recvice.Flush(); err != nil {
				ctx.Close(ctx.Self())
				return
			}
		}

		msg, err := protomessge.UnMarshal(dl.recvice, dl.secret)
		if err != nil {
			vlog.Errorf("unmarshal message error:%v", err.Error())
			ctx.Close(ctx.Self())
			return
		}

		if msg == nil {
			if werr != nil {
				vlog.Errorf("unmarshal message error:%v", werr.Error())
				ctx.Close(ctx.Self())
				return
			}

			if offset == len(ctx.Message()) {
				return
			}
			continue
		}

		/*if err := dl.message_agent.Message(ctx, msg, 0); err != nil {
			vlog.Error(err.Error())
			ctx.Close(ctx.Self())
			return
		}*/
	}
}

func (dl *ClientConn) Closed(ctx network.Context) {
	request := &prvs.ClientClosed{
		ClientID: ctx.Self(),
	}
	r := dl.gateway.FindRouter("OnClientClosed")
	if r != nil {
		clent := prvs.NewClientClosedServiceClient(r)
		if err := clent.OnClientClosed(context.Background(), request); err != nil {
			vlog.Errorf("request client %s closed to %s", request.ClientID.ToString(), "")
		}
	}

	dl.gateway.UnRegister(ctx.Self()) //关闭释放对象
}

func (dl *ClientConn) Destory() {
	dl.clientID = nil
	// dl.gateway = nil
	dl.secret = nil
	dl.recvice.Release()
}

// onUpdateRule 更改角色等级信息
func (dl *ClientConn) alterRule(rule int32) {
	dl.ruleID = rule
}

// RefInc 引用计数器+1
func (dl *ClientConn) referenceIncrement() int32 {
	return atomic.AddInt32(&dl.reference, 1)
}

// RefDec 引用计数器-1
func (dl *ClientConn) referenceDecrement() int32 {
	return atomic.AddInt32(&dl.reference, -1)
}
