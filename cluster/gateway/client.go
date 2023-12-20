package gateway

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"sync/atomic"

	"github.com/google/uuid"
	protomessge "github.com/yamakiller/velcro-go/cluster/gateway/protomessage"
	"github.com/yamakiller/velcro-go/cluster/protocols"
	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/circbuf"
)

type Client interface {
	ClientID() *network.ClientID
	Accept(ctx network.Context)
	Ping(ctx network.Context)
	Recvice(ctx network.Context)
	Closed(ctx network.Context)
	Destory()

	onPingReply(ctx network.Context, message interface{})
	onPubkeyReply(ctx network.Context, message interface{})
	onOtherMessage(ctx network.Context, message interface{})

	referenceIncrement() int32
	referenceDecrement() int32
}

type ClientConn struct {
	gateway   *Gateway
	clientID  *network.ClientID
	ruleID    int32  //角色ID
	secret    []byte //密钥
	recvice   *circbuf.RingBuffer
	ping      uint64
	reference int32 //引用计数器
}

func (dl *ClientConn) ClientID() *network.ClientID {
	return dl.clientID
}

func (dl *ClientConn) Accept(ctx network.Context) {
	dl.clientID = ctx.Self()
	dl.ruleID = router.NONE_RULE_ID
	dl.gateway.Register(dl.clientID, dl)
}

func (dl *ClientConn) Ping(ctx network.Context) {
	if dl.ruleID <= router.KEYED_RULE_ID {
		return
	}

	uid := uuid.New()
	dl.ping, _ = strconv.ParseUint(uid.String(), 10, 64)

	msg := &protocols.PingMsg{VerificationKey: dl.ping}

	var temp [128]byte
	msgLen, err := protomessge.Marshal(temp[:], msg, dl.secret)
	if err != nil {
		ctx.Error("ping marshal message [error:%v]", err.Error())
		return
	}

	ctx.PostMessage(ctx.Self(), temp[:msgLen])
}

func (dl *ClientConn) Recvice(ctx network.Context) {
	offset := 0
	for {
		n, _ := dl.recvice.Write(ctx.Message()[offset:])
		offset += n

		msg, err := protomessge.UnMarshal(dl.recvice, dl.secret)
		if err != nil {
			ctx.Error("unmarshal message error:%v", err.Error())
			ctx.Close(ctx.Self())
			return
		}

		switch message := msg.(type) {
		case *protocols.PingMsg:
			dl.onPingReply(ctx, message)
		case *protocols.PubkeyMsg:
			dl.onPubkeyReply(ctx, message)
		default:
			dl.onOtherMessage(ctx, message)
		}
	}

}

func (dl *ClientConn) Closed(ctx network.Context) {
	dl.gateway.UnRegister(ctx.Self()) //关闭释放对象
}

func (dl *ClientConn) Destory() {
	dl.clientID = nil
	dl.gateway = nil

	dl.recvice.Reset()
}

func (dl *ClientConn) onPingReply(ctx network.Context, message interface{}) {
	pingReq := message.(*protocols.PingMsg)
	if dl.ping == 0 {
		ctx.Debug("unrequest ping")
		ctx.Close(ctx.Self())
		return
	}

	if dl.ping+1 != pingReq.VerificationKey {
		ctx.Debug("ping reply error %d/%d", dl.ping+1, pingReq.VerificationKey)
		ctx.Close(ctx.Self())
		return
	}

	dl.ping = 0
	ctx.Debug("ping reply success")
}

func (dl *ClientConn) onPubkeyReply(ctx network.Context, message interface{}) {
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

	pubkeyReq := message.(*protocols.PubkeyMsg)

	var (
		prvKey crypto.PrivateKey
		pubKey crypto.PublicKey
	)

	pubkeyByte, err := base64.StdEncoding.DecodeString(pubkeyReq.Key)
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
	var tmp [256]byte
	pubkeyMessage := &protocols.PubkeyMsg{Key: base64.StdEncoding.EncodeToString(dl.gateway.encryption.Ecdh.Marshal(pubKey))}
	n, err := protomessge.Marshal(tmp[:], pubkeyMessage, nil)
	if err != nil {
		ctx.Debug("marshal pubkey message error %s", err.Error())
		ctx.Close(ctx.Self())
		return
	}

	ctx.PostMessage(ctx.Self(), tmp[:n])
}

func (dl *ClientConn) onOtherMessage(ctx network.Context, message interface{}) {
	// TODO: 编写其它处理代码
}

// RefInc 引用计数器+1
func (dl *ClientConn) referenceIncrement() int32 {
	return atomic.AddInt32(&dl.reference, 1)
}

// RefDec 引用计数器-1
func (dl *ClientConn) referenceDecrement() int32 {
	return atomic.AddInt32(&dl.reference, -1)
}
