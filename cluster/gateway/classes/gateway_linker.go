package classes

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/yamakiller/velcro-go/cluster/gateway/protocols"
	protomessge "github.com/yamakiller/velcro-go/cluster/gateway/protocols/protomessage"
	"github.com/yamakiller/velcro-go/cluster/route/router"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/utils/circbuf"
)

type GLinker interface {
	ClientID() *network.ClientID
	Accept(ctx network.Context)
	Ping(ctx network.Context)
	Recvice(ctx network.Context)
	Closed(ctx network.Context)
	Destory()

	onPingReply(ctx network.Context, message interface{})
	onPubkeyReply(ctx network.Context, message interface{})

	referenceIncrement() int32
	referenceDecrement() int32
}

type Linker struct {
	gateway      *Gateway
	clientID     *network.ClientID
	ruleID       int32  //角色ID
	secret       []byte //密钥
	recvice      *circbuf.RingBuffer
	ping         uint64
	otherMessage func(network.Context, interface{})
	reference    int32 //引用计数器
}

func (dl *Linker) ClientID() *network.ClientID {
	return dl.clientID
}

func (dl *Linker) Accept(ctx network.Context) {
	dl.clientID = ctx.Self()
	dl.ruleID = router.NONE_RULE_ID
	dl.gateway.linkerGroup.Register(dl.clientID, dl)
}

func (dl *Linker) Ping(ctx network.Context) {
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

func (dl *Linker) Recvice(ctx network.Context) {
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
			dl.otherMessage(ctx, message)
		}
	}

}

func (dl *Linker) Closed(ctx network.Context) {
	dl.gateway.linkerGroup.UnRegister(ctx.Self()) //关闭释放对象
}

func (dl *Linker) Destory() {
	dl.clientID = nil
	dl.secret = nil
	dl.gateway = nil
	dl.recvice = nil
}

func (dl *Linker) onPingReply(ctx network.Context, message interface{}) {
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

func (dl *Linker) onPubkeyReply(ctx network.Context, message interface{}) {
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

// RefInc 引用计数器+1
func (dl *Linker) referenceIncrement() int32 {
	return atomic.AddInt32(&dl.reference, 1)
}

// RefDec 引用计数器-1
func (dl *Linker) referenceDecrement() int32 {
	return atomic.AddInt32(&dl.reference, -1)
}
