package core

import (
	"fmt"

	"github.com/yamakiller/velcro-go/behavior"
)

type IBaseWorker interface {
	// 进入方法(可重载): 每次要求节点执行时，在tick本身之前都会调用它.
	OnEnter(tick *Tick)
	// 打开方法(可重载): 仅在tick回调之前并且仅当not未关闭时才调用它.
	OnOpen(tick *Tick)
	// TICK方法(可重载): 该方法必须包含节点的实际执行(执行任务、调用子级等).每次要求节点执行时都会调用它.
	OnTick(tick *Tick) behavior.Status

	// 关闭方法(可重载): 仅当状态返回与“RUNNING”不同的状态时, 才会在状态回调之后调用此方法.
	OnClose(tick *Tick)
	// 退出方法(可重载)
	OnExit(tick *Tick)
}

type BaseWorker struct {
}

// 进入方法
func (this *BaseWorker) OnEnter(tick *Tick) {

}

// 打开方法
func (this *BaseWorker) OnOpen(tick *Tick) {

}

// TICK方法
func (this *BaseWorker) OnTick(tick *Tick) behavior.Status {
	fmt.Println("tick BaseWorker")
	return behavior.ERROR
}

// 关闭方法
func (this *BaseWorker) OnClose(tick *Tick) {

}

// 退出方法: 覆盖它来使用.每次执行结束时都会调用.
func (this *BaseWorker) OnExit(tick *Tick) {

}
