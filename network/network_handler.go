package network

import "sync/atomic"

type SpawnHandler func(system *NetworkSystem, id string) (*ClientId, error)

type Handler interface {
	PostSysMessage(*ClientId, interface{})
	PostUsrMessage(*ClientId, interface{})
	Close(*ClientId)
}

type ClientHandler struct {
	_dead int32
}

func (chr *ClientHandler) PostSysMessage(_ *ClientId, message interface{}) {

}

func (chr *ClientHandler) PostUsrMessage(_ *ClientId, message interface{}) {

}

func (chr *ClientHandler) Close(cid *ClientId) {
	atomic.StoreInt32(&chr._dead, 1)
	// 发送关闭消息
	chr.PostSysMessage(cid, closeMessage)
}
