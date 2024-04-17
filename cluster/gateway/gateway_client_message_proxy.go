package gateway

import (
	"context"
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/apache/thrift/lib/go/thrift"
	protomessge "github.com/yamakiller/velcro-go/cluster/gateway/protomessage"
	"github.com/yamakiller/velcro-go/cluster/protocols/prvs"
	"github.com/yamakiller/velcro-go/cluster/protocols/pubs"
	"github.com/yamakiller/velcro-go/cluster/router"
	"github.com/yamakiller/velcro-go/network"
	"github.com/yamakiller/velcro-go/rpc/messages"
	"github.com/yamakiller/velcro-go/rpc/protocol"
	"github.com/yamakiller/velcro-go/cluster/proxy/messageproxy"
)

func NewPubkeyMessageProxy(dl *ClientConn)*PubkeyMessageProxy{
	return &PubkeyMessageProxy{
		IMessageProxyNode: messageproxy.NewMessageProxyNode(),
		dl: dl,
	}
}

type PubkeyMessageProxy struct {
	messageproxy.IMessageProxyNode
	dl *ClientConn
}

func (p *PubkeyMessageProxy) Message(ctx network.Context, msg []byte,timeout int64) error {
	iprot := protocol.NewBinaryProtocol()
	defer iprot.Close()
	iprot.Write(msg)
	_, _, seqid, err := iprot.ReadMessageBegin(context.Background())
	if err != nil {
		return err
	}
	message := pubs.NewPubkeyMsg()
	if err := message.Read(context.Background(), iprot); err != nil {
		return fmt.Errorf("msg is net PubkeyMsg")
	}
	if err := iprot.ReadMessageEnd(context.Background()); err != nil {
		return err
	}
	err = p.onPubkeyReply(ctx, message, iprot, seqid)
	p.Next()
	return nil
}

func (p *PubkeyMessageProxy) onPubkeyReply(ctx network.Context, message *pubs.PubkeyMsg, iprot protocol.IProtocol, seqId int32) error {
	if p.dl.gateway.encryption == nil {
		return errors.New("encrypted communication not enabled")
	}

	if p.dl.ruleID != router.NONE_RULE_ID {
		return errors.New("key exchange completed, abnormal process")
	}

	var (
		prvKey crypto.PrivateKey
		pubKey crypto.PublicKey
	)

	pubkeyByte, err := base64.StdEncoding.DecodeString(message.Key)
	if err != nil {
		return fmt.Errorf("public key decode error %s", err.Error())
	}

	prvKey, pubKey, err = p.dl.gateway.encryption.Ecdh.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate public/private key error %s", err.Error())
	}

	remotePubkey, ok := p.dl.gateway.encryption.Ecdh.Unmarshal(pubkeyByte)
	if !ok {
		return fmt.Errorf("Public key parsing exception")
	}

	secret, err := p.dl.gateway.encryption.Ecdh.GenerateSharedSecret(prvKey, remotePubkey)
	if err != nil {
		return fmt.Errorf("generate shared secret error %s", err.Error())
	}

	// 密钥生成结束
	p.dl.secret = secret
	p.dl.ruleID = router.KEYED_RULE_ID

	// 回复消息

	message.Key = base64.StdEncoding.EncodeToString(p.dl.gateway.encryption.Ecdh.Marshal(pubKey))
	if err != nil {
		return fmt.Errorf("marshal pubkey message error %s", err.Error())
	}
	b, err := messages.MarshalTStruct(context.Background(), iprot, message, protocol.MessageName(message), seqId)
	if err != nil {
		return fmt.Errorf("MarshalTStruct pubkey message error %s", err.Error())
	}
	ctx.PostMessage(ctx.Self(), b)
	return nil
}

func NewPingMessageProxy(dl *ClientConn) *PingMessageProxy{
	return &PingMessageProxy{
		IMessageProxyNode: messageproxy.NewMessageProxyNode(),
		dl: dl,
		message: pubs.NewPingMsg(),
	}
}

type PingMessageProxy struct{
	messageproxy.IMessageProxyNode
	dl *ClientConn
	message *pubs.PingMsg
}
func (pp *PingMessageProxy) UnMarshal(msg []byte)error{
	iprot := protocol.NewBinaryProtocol()
	defer iprot.Close()
	iprot.Write(msg)
	if _,_,_,err := iprot.ReadMessageBegin(context.Background());err != nil{
		return err
	}
	if err := pp.message.Read(context.Background(),iprot); err != nil{
		return err
	}
	if err := iprot.ReadMessageEnd(context.Background()); err != nil{
		return err
	}
	return nil
}

func (pp *PingMessageProxy) Method(ctx network.Context, seqId int32,timeout int64) error{
	iprot := protocol.NewBinaryProtocol()
	defer iprot.Close()
	if pp.dl.ping == 0 {
		return fmt.Errorf("unrequest ping")
	}

	if pp.dl.ping+1 != uint64(pp.message.VerificationKey) {
		return fmt.Errorf("ping reply error %d/%d",pp. dl.ping+1, uint64(pp.message.VerificationKey) )
	}
	pp.dl.ping = 0
	return nil
}


func NewRequestMessageProxy(dl *ClientConn) *RequestMessageProxy {
	return &RequestMessageProxy{
		IMessageProxyNode: messageproxy.NewMessageProxyNode(),
		dl:                dl,
		message: pubs.NewRequestMessage(),
	}
}

type RequestMessageProxy struct {
	messageproxy.IMessageProxyNode
	dl *ClientConn
	message *pubs.RequestMessage
}

func (rp *RequestMessageProxy) UnMarshal(msg []byte)error{
	iprot := protocol.NewBinaryProtocol()
	defer iprot.Close()
	iprot.Write(msg)
	if _,_,_,err := iprot.ReadMessageBegin(context.Background());err != nil{
		return err
	}
	if err := rp.message.Read(context.Background(),iprot); err != nil{
		return err
	}
	if err := iprot.ReadMessageEnd(context.Background()); err != nil{
		return err
	}
	return nil
}

func (rp *RequestMessageProxy) Method(ctx network.Context, seqId int32,timeout int64)error {
	iprot := protocol.NewBinaryProtocol()
	defer iprot.Close()
	iprot.Write(rp.message.Msg)
	name ,_,seq,err :=  iprot.ReadMessageBegin(context.Background())
	if err != nil{
		return err
	}
	r := rp.dl.gateway.FindRouter(name)
	if r == nil {
		return fmt.Errorf("%s message unfound router",
		name)
	}
	if !r.IsRulePass(rp.dl.ruleID) {
		return fmt.Errorf("%s message Insufficient permissions",
		name)
	}
	forwardBundle := &prvs.ForwardBundle{
		Sender: rp.dl.clientID,
		Body:   rp.message.Msg,
	}
	iprot.Release()
	// 采用平均时间
	result, err := r.Proxy.RequestMessage(forwardBundle, protocol.MessageName(forwardBundle), r.GetMessageTimeout(name))
	if err != nil {
		er := &pubs.Error{
			Name: name,
			Err:  err.Error(),
		}
		iprot.WriteMessageBegin(context.Background(), protocol.MessageName(er), thrift.EXCEPTION, seq)
		er.Write(context.Background(), iprot)
		b, msge := protomessge.Marshal(iprot.GetBytes(), rp.dl.secret)
		if msge != nil {
			panic(msge)
		}

		ctx.PostMessage(ctx.Self(), b)
		return nil
	}
	if result == nil {
		return fmt.Errorf("message %s result is nil",name)
	}

	b, msge := protomessge.Marshal(result,rp.dl.secret)
	if msge != nil {
		return fmt.Errorf("requesting pubs.Error marshal %s message fail[error:%s]", name, msge.Error())
	}
	ctx.PostMessage(ctx.Self(), b)
	return nil
}
