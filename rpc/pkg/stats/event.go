package stats

import (
	"errors"
	"sync"
	"sync/atomic"
)

// EventIndex 表示一个独特的事件
type EventIndex int

// Level 设置记录等级.
type Level int

// Event levels.
const (
	LevelDisabled Level = iota
	LevelBase
	LevelDetailed
)

// Event 用于指示特定事件.
type Event interface {
	Index() EventIndex
	Level() Level
}

type event struct {
	idx   EventIndex
	level Level
}

// Index implements the Event interface.
func (e event) Index() EventIndex {
	return e.idx
}

// Level implements the Event interface.
func (e event) Level() Level {
	return e.level
}

const (
	_ EventIndex = iota
	serverHandleStart
	serverHandleFinish
	clientConnStart
	clientConnFinish
	rpcStart
	rpcFinish
	readStart
	readFinish
	waitReadStart
	waitReadFinish
	writeStart
	writeFinish
	streamRecv
	streamSend

	// NOTE: add new events before this line
	predefinedEventNum
)

// 预定义的事件 events.
var (
	RPCStart  = newEvent(rpcStart, LevelBase)
	RPCFinish = newEvent(rpcFinish, LevelBase)

	ServerHandleStart  = newEvent(serverHandleStart, LevelDetailed)
	ServerHandleFinish = newEvent(serverHandleFinish, LevelDetailed)
	ClientConnStart    = newEvent(clientConnStart, LevelDetailed)
	ClientConnFinish   = newEvent(clientConnFinish, LevelDetailed)
	ReadStart          = newEvent(readStart, LevelDetailed)
	ReadFinish         = newEvent(readFinish, LevelDetailed)
	WaitReadStart      = newEvent(waitReadStart, LevelDetailed)
	WaitReadFinish     = newEvent(waitReadFinish, LevelDetailed)
	WriteStart         = newEvent(writeStart, LevelDetailed)
	WriteFinish        = newEvent(writeFinish, LevelDetailed)

	// Streaming Events
	StreamRecv = newEvent(streamRecv, LevelDetailed)
	StreamSend = newEvent(streamSend, LevelDetailed)
)

// errors
var (
	ErrNotAllowed = errors.New("event definition is not allowed after initialization")
	ErrDuplicated = errors.New("event name is already defined")
)

var (
	lock        sync.RWMutex
	inited      int32
	userDefined = make(map[string]Event)
	maxEventNum = int(predefinedEventNum)
)

// FinishInitialization 冻结所有定义的事件并阻止添加进一步的定义.
func FinishInitialization() {
	lock.Lock()
	defer lock.Unlock()
	atomic.StoreInt32(&inited, 1)
}

// DefineNewEvent 允许用户在程序初始化期间添加事件定义
func DefineNewEvent(name string, level Level) (Event, error) {
	if atomic.LoadInt32(&inited) == 1 {
		return nil, ErrNotAllowed
	}
	lock.Lock()
	defer lock.Unlock()
	evt, exist := userDefined[name]
	if exist {
		return evt, ErrDuplicated
	}
	userDefined[name] = newEvent(EventIndex(maxEventNum), level)
	maxEventNum++
	return userDefined[name], nil
}

// MaxEventNum 返回定义的事件数
func MaxEventNum() int {
	lock.RLock()
	defer lock.RUnlock()
	return maxEventNum
}

// PredefinedEventNum 返回 velcro 预定义事件的数量.
func PredefinedEventNum() int {
	return int(predefinedEventNum)
}

func newEvent(idx EventIndex, level Level) Event {
	return event{
		idx:   idx,
		level: level,
	}
}
