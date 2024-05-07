package event

import (
	"context"
	"reflect"
	"sync"

	"github.com/yamakiller/velcro-go/gofunc"
)

// Callback 事件回调函数
type Callback func(*Event)

// Bus 允许事件调度和观察的观察者模式.
type Bus interface {
	Watch(event string, callback Callback)
	Unwatch(event string, callback Callback)
	Dispatch(event *Event)
	DispatchAndWait(event *Event)
}

// NewEventBus 创建一个总线
func NewEventBus() Bus {
	return &bus{}
}

type bus struct {
	callbacks sync.Map
}

// Watch 通过回调订阅特定事件.
func (b *bus) Watch(event string, callback Callback) {
	var callbacks []Callback
	if actual, ok := b.callbacks.Load(event); ok {
		callbacks = append(callbacks, actual.([]Callback)...)
	}
	callbacks = append(callbacks, callback)
	b.callbacks.Store(event, callbacks)
}

// Unwatch 从事件的回调链中删除给定的回调.
func (b *bus) Unwatch(event string, callback Callback) {
	var filtered []Callback
	// 使用reflect.ValueOf(callback).Pointer()获取函数地址进行比较,两个函数是否相同.
	target := reflect.ValueOf(callback).Pointer()
	if actual, ok := b.callbacks.Load(event); ok {
		for _, h := range actual.([]Callback) {
			if reflect.ValueOf(h).Pointer() != target {
				filtered = append(filtered, h)
			}
		}
		b.callbacks.Store(event, filtered)
	}
}

// Dispatch 过异步调用每个回调来调度事件.
func (b *bus) Dispatch(event *Event) {
	if actual, ok := b.callbacks.Load(event.Name); ok {
		for _, h := range actual.([]Callback) {
			f := h // 将值分配给闭包的新变量
			gofunc.GoFunc(context.Background(), func() {
				f(event)
			})
		}
	}
}

// DispatchAndWait dispatches an event by invoking callbacks concurrently and waits for them to finish.
func (b *bus) DispatchAndWait(event *Event) {
	if actual, ok := b.callbacks.Load(event.Name); ok {
		var wg sync.WaitGroup
		for i := range actual.([]Callback) {
			h := (actual.([]Callback))[i]
			wg.Add(1)
			gofunc.GoFunc(context.Background(), func() {
				h(event)
				wg.Done()
			})
		}
		wg.Wait()
	}
}
