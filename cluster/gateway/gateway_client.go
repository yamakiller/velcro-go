package gateway

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"math"
	"reflect"
	"sync/atomic"

	protomessge "github.com/yamakiller/velcro-go/cluster/gateway/protomessage"
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/protocols/pubs"
	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"github.com/yamakiller/velcro-go/utils/lang/fastrand"
	"github.com/yamakiller/velcro-go/vlog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type Client interface {
	ClientID() *network.ClientID
	Secret() []byte
	Accept(ctx network.Context)
	Ping(ctx network.Context)
	Post(message interface{}) error
	Recvice(ctx network.Context)
	Closed(ctx network.Context)
	Destory()

	alterRule(rule int32)

	onPingReply(ctx network.Context, message *pubs.PingMsg)
	onPubkeyReply(ctx network.Context, message *pubs.PubkeyMsg)
	onRequestMessage(ctx network.Context, message proto.Message)

	referenceIncrement() int32
	referenceDecrement() int32
}

type ClientConn struct {
	gateway  *Gateway
	clientID *network.ClientID
	ruleID   int32  //角色ID
	secret   []byte //密钥
	recvice  *circbuf.LinkBuffer
	ping     uint64
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

	dl.ping = fastrand.Uint64n(math.MaxUint64)

	msg := &pubs.PingMsg{VerificationKey: dl.ping}
	msgb, err := protomessge.Marshal(msg, dl.secret)
	if err != nil {
		vlog.Errorf("ping marshal message [error:%v]", err.Error())
		return
	}

	ctx.PostMessage(ctx.Self(), msgb)
}

func (dl *ClientConn) Post(message interface{}) error {
	b, err := protomessge.Marshal(message.(proto.Message), dl.secret)
	if err != nil {
		return err
	}

	return dl.clientID.PostUserMessage(b)
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
				vlog.Errorf("unmarshal message error:%v", err.Error())
				ctx.Close(ctx.Self())
				return
			}

			if offset == len(ctx.Message()) {
				return
			}
			continue
		}

		switch message := msg.(type) {
		case *pubs.PingMsg:
			dl.onPingReply(ctx, message)
		case *pubs.PubkeyMsg:
			dl.onPubkeyReply(ctx, message)
		default:
			dl.onRequestMessage(ctx, message)
		}
	}
}

func (dl *ClientConn) Closed(ctx network.Context) {
	request := &prvs.ClientClosed{
		ClientID: ctx.Self(),
	}

	r := dl.gateway.FindRouter(request)
	if r != nil {
		if _, err := r.Proxy.RequestMessage(request, r.GetMessageTimeout(request)); err != nil {
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

func (dl *ClientConn) onPingReply(ctx network.Context, message *pubs.PingMsg) {

	if dl.ping == 0 {
		vlog.Debug("unrequest ping")
		ctx.Close(ctx.Self())
		return
	}

	if dl.ping+1 != message.VerificationKey {
		vlog.Debugf("ping reply error %d/%d", dl.ping+1, message.VerificationKey)
		ctx.Close(ctx.Self())
		return
	}

	dl.ping = 0
}

func (dl *ClientConn) onPubkeyReply(ctx network.Context, message *pubs.PubkeyMsg) {
	if dl.gateway.encryption == nil {
		vlog.Debug("encrypted communication not enabled")
		ctx.Close(ctx.Self())
		return
	}

	if dl.ruleID != router.NONE_RULE_ID {
		vlog.Debug("key exchange completed, abnormal process")
		ctx.Close(ctx.Self())
		return
	}

	var (
		prvKey crypto.PrivateKey
		pubKey crypto.PublicKey
	)

	pubkeyByte, err := base64.StdEncoding.DecodeString(message.Key)
	if err != nil {
		vlog.Debugf("public key decode error %s", err.Error())
		ctx.Close(ctx.Self())
		return
	}

	prvKey, pubKey, err = dl.gateway.encryption.Ecdh.GenerateKey(rand.Reader)
	if err != nil {
		vlog.Debugf("generate public/private key error %s", err.Error())
		ctx.Close(ctx.Self())
	}

	remotePubkey, ok := dl.gateway.encryption.Ecdh.Unmarshal(pubkeyByte)
	if !ok {
		vlog.Debug("Public key parsing exception")
		ctx.Close(ctx.Self())
		return
	}

	secret, err := dl.gateway.encryption.Ecdh.GenerateSharedSecret(prvKey, remotePubkey)
	if err != nil {
		vlog.Debugf("generate shared secret error %s", err.Error())
		ctx.Close(ctx.Self())
		return
	}

	// 密钥生成结束
	dl.secret = secret
	dl.ruleID = router.KEYED_RULE_ID

	// 回复消息
	pubkeyMessage := &pubs.PubkeyMsg{Key: base64.StdEncoding.EncodeToString(dl.gateway.encryption.Ecdh.Marshal(pubKey))}
	b, err := protomessge.Marshal(pubkeyMessage, nil)
	if err != nil {
		vlog.Debugf("marshal pubkey message error %s", err.Error())
		ctx.Close(ctx.Self())
		return
	}

	ctx.PostMessage(ctx.Self(), b)
}

func (dl *ClientConn) onRequestMessage(ctx network.Context, message proto.Message) {
	r := dl.gateway.FindRouter(message)
	if r == nil {
		vlog.Warnf("%s message unfound router",
			string(protoreflect.FullName(proto.MessageName(message))))
		ctx.Close(ctx.Self())
		return
	}

	if !r.IsRulePass(dl.ruleID) {
		vlog.Warnf("%s message Insufficient permissions",
			string(protoreflect.FullName(proto.MessageName(message))))
		ctx.Close(ctx.Self())
		return
	}

	bodyAny, err := anypb.New(message)
	if err != nil {
		vlog.Warnf("%s message encoding failed error %s",
			string(protoreflect.FullName(proto.MessageName(message))), err.Error())
		ctx.Close(ctx.Self())
		return
	}

	forwardBundle := &prvs.ForwardBundle{
		Sender: dl.clientID,
		Body:   bodyAny,
	}

	// 采用平均时间
	result, err := r.Proxy.RequestMessage(forwardBundle, r.GetMessageTimeout(message))
	if err != nil {

		b, msge := protomessge.Marshal(&pubs.Error{
			Name: string(protoreflect.FullName(proto.MessageName(message))),
			Err:  err.Error(),
		}, dl.secret)

		if msge != nil {
			panic(msge)
		}

		ctx.PostMessage(ctx.Self(), b)
		return
	}
	if result == nil {
		return
	}
	b, msge := protomessge.Marshal(result, dl.secret)
	if msge != nil {
		vlog.Errorf("requesting pubs.Error marshal %s message fail[error:%s]", reflect.TypeOf(result).Name(), msge.Error())
		ctx.Close(ctx.Self())
		return
	}

	ctx.PostMessage(ctx.Self(), b)
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
