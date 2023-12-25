package gateway

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	protomessge "github.com/yamakiller/velcro-go/cluster/gateway/protomessage"
	"github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/utils/circbuf"
	"google.golang.org/protobuf/proto"
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

	onPingReply(ctx network.Context, message *protocols.PingMsg)
	onPubkeyReply(ctx network.Context, message *protocols.PubkeyMsg)
	onRequestMessage(ctx network.Context, message *protocols.ClientRequestMessage)
	onPostMessage(ctx network.Context, message proto.Message)
	onUpdateRule(rule int32)

	referenceIncrement() int32
	referenceDecrement() int32
}

type ClientConn struct {
	gateway             *Gateway
	clientID            *network.ClientID
	ruleID              int32  //角色ID
	secret              []byte //密钥
	recvice             *circbuf.RingBuffer
	ping                uint64
	message_max_timeout int64 //最大超时时间 毫秒级
	reference           int32 //引用计数器
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

	uid := uuid.New()
	dl.ping, _ = strconv.ParseUint(uid.String(), 10, 64)

	msg := &protocols.PingMsg{VerificationKey: dl.ping}
	msgb, err := protomessge.Marshal(msg, dl.secret)
	if err != nil {
		ctx.Error("ping marshal message [error:%v]", err.Error())
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
			n, werr = dl.recvice.Write(ctx.Message()[offset:])
			offset += n
		}

		msg, err := protomessge.UnMarshal(dl.recvice, dl.secret)
		if err != nil {
			ctx.Error("unmarshal message error:%v", err.Error())
			ctx.Close(ctx.Self())
			return
		}

		if msg == nil {
			if werr != nil {
				ctx.Error("unmarshal message error:%v", err.Error())
				ctx.Close(ctx.Self())
				return
			}

			if offset == len(ctx.Message()) {
				return
			}
			continue
		}

		switch message := msg.(type) {
		case *protocols.PingMsg:
			dl.onPingReply(ctx, message)
		case *protocols.PubkeyMsg:
			dl.onPubkeyReply(ctx, message)
		case *protocols.ClientRequestMessage:
			dl.onRequestMessage(ctx, message)
		default:
			dl.onPostMessage(ctx, message)
		}
	}

}

func (dl *ClientConn) Closed(ctx network.Context) {
	dl.gateway.UnRegister(ctx.Self()) //关闭释放对象
}

func (dl *ClientConn) Destory() {
	dl.clientID = nil
	dl.gateway = nil
	dl.secret = nil
	dl.recvice.Reset()
}

func (dl *ClientConn) onPingReply(ctx network.Context, message *protocols.PingMsg) {

	if dl.ping == 0 {
		ctx.Debug("unrequest ping")
		ctx.Close(ctx.Self())
		return
	}

	if dl.ping+1 != message.VerificationKey {
		ctx.Debug("ping reply error %d/%d", dl.ping+1, message.VerificationKey)
		ctx.Close(ctx.Self())
		return
	}

	dl.ping = 0
	ctx.Debug("ping reply success")
}

func (dl *ClientConn) onPubkeyReply(ctx network.Context, message *protocols.PubkeyMsg) {
	if dl.gateway.encryption == nil {
		ctx.Debug("encrypted communication not enabled")
		ctx.Close(ctx.Self())
		return
	}

	if dl.ruleID != router.NONE_RULE_ID {
		ctx.Debug("key exchange completed, abnormal process")
		ctx.Close(ctx.Self())
		return
	}

	var (
		prvKey crypto.PrivateKey
		pubKey crypto.PublicKey
	)

	pubkeyByte, err := base64.StdEncoding.DecodeString(message.Key)
	if err != nil {
		ctx.Debug("public key decode error %s", err.Error())
		ctx.Close(ctx.Self())
		return
	}

	prvKey, pubKey, err = dl.gateway.encryption.Ecdh.GenerateKey(rand.Reader)
	if err != nil {
		ctx.Debug("generate public/private key error %s", err.Error())
		ctx.Close(ctx.Self())
	}

	remotePubkey, ok := dl.gateway.encryption.Ecdh.Unmarshal(pubkeyByte)
	if !ok {
		ctx.Debug("Public key parsing exception")
		ctx.Close(ctx.Self())
		return
	}

	secret, err := dl.gateway.encryption.Ecdh.GenerateSharedSecret(prvKey, remotePubkey)
	if err != nil {
		ctx.Debug("generate shared secret error %s", err.Error())
		ctx.Close(ctx.Self())
		return
	}

	// 密钥生成结束
	dl.secret = secret
	dl.ruleID = router.KEYED_RULE_ID

	// 回复消息
	pubkeyMessage := &protocols.PubkeyMsg{Key: base64.StdEncoding.EncodeToString(dl.gateway.encryption.Ecdh.Marshal(pubKey))}
	b, err := protomessge.Marshal(pubkeyMessage, nil)
	if err != nil {
		ctx.Debug("marshal pubkey message error %s", err.Error())
		ctx.Close(ctx.Self())
		return
	}

	ctx.PostMessage(ctx.Self(), b)
}

func (dl *ClientConn) onRequestMessage(ctx network.Context, message *protocols.ClientRequestMessage) {
	requestMessageName := string(message.RequestMessage.MessageName())
	requestMessage := message.RequestMessage.ProtoReflect().New().Interface()
	if err := message.RequestMessage.UnmarshalTo(requestMessage); err != nil {
		ctx.Error("requesting unmarshal fail[error:%s]", requestMessageName, err.Error())
		ctx.Close(ctx.Self())
		return
	}

	r := dl.gateway.routeGroup.Get(requestMessageName)
	if r != nil {
		ctx.Warning("%s message unfound router", requestMessageName)
		ctx.Close(ctx.Self())
		return
	}

	if !r.IsRulePass(dl.ruleID) {
		ctx.Warning("%s message Insufficient permissions", requestMessageName)
		ctx.Close(ctx.Self())
		return
	}

	//TODO: 优化时间计算
	if int64(message.RequestTimeout) > dl.message_max_timeout {
		ctx.Warning("%s message timeout to long ,max timeout is %d", requestMessageName, dl.message_max_timeout)
		return
	}
	timeout := int64(message.RequestTimeout) - (time.Now().UnixMilli() - int64(message.RequestTime))
	if timeout <= 0 {
		ctx.Warning("%s message timeout", requestMessageName)
		return
	}
	result, err := r.Proxy.RequestMessage(message, int64(message.RequestTimeout))
	if err != nil {
		b, msge := protomessge.Marshal(&protocols.Error{
			ID:   message.RequestID,
			Name: requestMessageName,
			Err:  err.Error(),
		}, dl.secret)
		if msge != nil {
			ctx.Error("requesting protocols.Error marshal %s message fail[error:%s]", requestMessageName, msge.Error())
			ctx.Close(ctx.Self())
			return
		}

		ctx.PostMessage(ctx.Self(), b)
		return
	}

	b, msge := protomessge.Marshal(result.(proto.Message), dl.secret)
	if msge != nil {
		ctx.Error("requesting protocols.Error marshal %s message fail[error:%s]", reflect.TypeOf(result).Name(), msge.Error())
		ctx.Close(ctx.Self())
		return
	}

	ctx.PostMessage(ctx.Self(), b)
}

func (dl *ClientConn) onPostMessage(ctx network.Context, message proto.Message) {

	msgName := proto.MessageName(message)
	r := dl.gateway.routeGroup.Get(string(msgName))
	if r != nil {
		ctx.Warning("%s message unfound router", msgName)
		ctx.Close(ctx.Self())
		return
	}

	if !r.IsRulePass(dl.ruleID) {
		ctx.Warning("%s message Insufficient permissions", msgName)
		ctx.Close(ctx.Self())
		return
	}

	err := r.Proxy.PostMessage(message, messages.RpcQosDiscard)
	if err != nil {
		b, msge := protomessge.Marshal(&protocols.Error{
			Name: string(msgName),
			Err:  err.Error(),
		}, dl.secret)
		if msge != nil {
			ctx.Error("posting protocols.Error marshal %s message fail[error:%s]", string(msgName), msge.Error())
			ctx.Close(ctx.Self())
			return
		}

		ctx.PostMessage(ctx.Self(), b)
		return
	}

	ctx.Debug("post message %s", msgName)
}

// onUpdateRule 更改角色等级信息
func (dl *ClientConn) onUpdateRule(rule int32) {
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
