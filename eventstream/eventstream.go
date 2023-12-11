package eventstream

import (
	"sync"
	"sync/atomic"
)

// 定义订阅时必须传递的回调函数
type Handler func(interface{})

// Predicate 在转发给订阅者之前过滤消息的函数
type Predicate func(evt interface{}) bool

type EventStream struct {
	sync.RWMutex
	_subscriptions []*Subscription
	_counter       int32
}

func NewEventStream() *EventStream {
	es := &EventStream{
		_subscriptions: []*Subscription{},
	}

	return es
}

type Subscription struct {
	_id      int32
	_handler Handler
	_p       Predicate
	_active  uint32
}

func (es *EventStream) Subscribe(handler Handler) *Subscription {
	sub := &Subscription{
		_handler: handler,
		_active:  1,
	}

	es.Lock()
	defer es.Unlock()

	sub._id = es._counter
	es._counter++
	es._subscriptions = append(es._subscriptions, sub)

	return sub
}

// SubscribeWithPredicate 创建一个新的订阅值并设置一个过滤函数来传递给的消息
// 订阅者,它返回指向订阅值的指针
func (es *EventStream) SubscribeWithPredicate(handler Handler, p Predicate) *Subscription {
	sub := es.Subscribe(handler)
	sub._p = p

	return sub
}

// Unsubscribe 来自 EventStream 的给定订阅
func (es *EventStream) Unsubscribe(sub *Subscription) {
	if sub == nil {
		return
	}

	if sub.IsActive() {
		es.Lock()
		defer es.Unlock()

		if sub.Deactivate() {
			if es._counter == 0 {
				es._subscriptions = nil

				return
			}

			l := es._counter - 1
			es._subscriptions[sub._id] = es._subscriptions[l]
			es._subscriptions[sub._id]._id = sub._id
			es._subscriptions[l] = nil
			es._subscriptions = es._subscriptions[:l]
			es._counter--

			if es._counter == 0 {
				es._subscriptions = nil
			}
		}
	}
}

// Publish 发布一个事件
func (es *EventStream) Publish(evt interface{}) {
	subs := make([]*Subscription, 0, es.Length())
	es.RLock()
	for _, sub := range es._subscriptions {
		if sub.IsActive() {
			subs = append(subs, sub)
		}
	}
	es.RUnlock()

	for _, sub := range subs {
		// there is a subscription predicate and it didn't pass, return
		if sub._p != nil && !sub._p(evt) {
			continue
		}

		// finally here, lets execute our handler
		sub._handler(evt)
	}
}

// Length 事件数量
func (es *EventStream) Length() int32 {
	es.RLock()
	defer es.RUnlock()
	return es._counter
}

// Activate 订阅将其活动标志设置为 1,
// 如果订阅已处于活动状态, 则返回 false, 否则返回 true
func (s *Subscription) Activate() bool {
	return atomic.CompareAndSwapUint32(&s._active, 0, 1)
}

// Deactivate订阅将其活动标志设置为 0，
// 如果订阅已处于非活动状态,则返回 false, 否则返回 true
func (s *Subscription) Deactivate() bool {
	return atomic.CompareAndSwapUint32(&s._active, 1, 0)
}

// IsActive 是否处于活动状态
func (s *Subscription) IsActive() bool {
	return atomic.LoadUint32(&s._active) == 1
}
