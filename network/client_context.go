package network

import (
	"net"
	"sync/atomic"
)

// 通用客户端上下文
type clientContext struct {
	client      Client
	system      *NetworkSystem
	self        *ClientID
	message     []byte
	messageFrom net.Addr
	state       int32
}

func (ctx *clientContext) Client() Client {
	return ctx.client
}

func (ctx *clientContext) Self() *ClientID {
	return ctx.self
}

func (ctx *clientContext) NetworkSystem() *NetworkSystem {
	return ctx.system
}

// Message 返回当前接收到的消息
func (ctx *clientContext) Message() []byte {
	return ctx.message
}

func (ctx *clientContext) MessageFrom() net.Addr {
	return ctx.messageFrom
}

func (ctx *clientContext) PostMessage(cid *ClientID, message []byte) error {
	return cid.postMessage(ctx.system, message)
}

func (ctx *clientContext) PostToMessage(cid *ClientID, message []byte, target net.Addr) error {
	return cid.postToMessage(ctx.system, message, target)
}

// Close 关闭当前 Client
func (ctx *clientContext) Close(cid *ClientID) {
	cid.close(ctx.system)
}

func (ctx *clientContext) incarnateClient() {
	atomic.StoreInt32(&ctx.state, stateAccept)

	ctx.client = ctx.system.producer(ctx.system)
}

func (ctx *clientContext) invokerAccept() {
	ctx.self.ref(ctx.system)
	ctx.client.Accept(ctx)
}

// MessageInvoker 接口实现
func (ctx *clientContext) invokerRecvice(b []byte, addr net.Addr) {
	if atomic.LoadInt32(&ctx.state) == stateClosed {
		// 已经关闭
		return
	}

	ctx.message = b
	ctx.messageFrom = addr
	ctx.client.Recvice(ctx)
	ctx.messageFrom = nil
	ctx.message = nil
}

func (ctx *clientContext) invokerPing() {
	if atomic.LoadInt32(&ctx.state) == stateClosed {
		// 已关闭
		return
	}

	ctx.client.Ping(ctx)
}

func (ctx *clientContext) invokerClosed() {
	if atomic.LoadInt32(&ctx.state) == stateClosed {
		// 已关闭
		return
	}

	atomic.StoreInt32(&ctx.state, stateClosed)
	atomic.StoreInt32(&ctx.self.vaild, 1)
	ctx.system.handlers.Remove(ctx.self)
	ctx.client.Closed(ctx)
}

func (ctx *clientContext) invokerEscalateFailure(reason interface{}, message interface{}) {
}
