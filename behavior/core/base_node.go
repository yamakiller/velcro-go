package core

import (
	"github.com/yamakiller/velcro-go/behavior"
	"github.com/yamakiller/velcro-go/behavior/datas"
)

type IBaseWrapper interface {
	execute(tick *Tick) behavior.Status
	enter(tick *Tick)
	open(tick *Tick)
	tick(tick *Tick) behavior.Status
	close(tick *Tick)
	exit(tick *Tick)
}

type IBaseNode interface {
	IBaseWrapper

	Ctor()
	Initialize(data *datas.Behavior3Node)
	GetCategory() string
	Execute(tick *Tick) behavior.Status
	GetName() string
	GetTitle() string
	SetBaseNodeWorker(worker IBaseWorker)
	GetBaseNodeWorker() IBaseWorker
}

type BaseNode struct {
	IBaseWorker

	// 节点ID
	id string
	// 节点名称. 必须是唯一标识符，最好与类的名称相同. 您必须在原型中设置节点名称.
	name string
	// 节点说明,可选项
	description string
	// 节点类别:  be `COMPOSITE`, `DECORATOR`, `ACTION` or `CONDITION`
	// 这是自动定义的, 继承相应的类.
	category string
	// 节点标题
	title string
	// 描述节点属性的字典(键,值).对于在可视化编辑器中定义自定义变量很有用.
	// properties 属性
	properties map[string]interface{}
}

func (this *BaseNode) Ctor() {

}

func (this *BaseNode) SetName(name string) {
	this.name = name
}
func (this *BaseNode) SetTitle(name string) {
	this.name = name
}

func (this *BaseNode) SetBaseNodeWorker(worker IBaseWorker) {
	this.IBaseWorker = worker
}

func (this *BaseNode) GetBaseNodeWorker() IBaseWorker {
	return this.IBaseWorker
}

/**
 * Initialization method.
 * @method Initialize
 * @construCtor
**/
func (this *BaseNode) Initialize(data *datas.Behavior3Node) {

	this.description = ""
	this.properties = make(map[string]interface{})

}

func (this *BaseNode) GetCategory() string {
	return this.category
}

func (this *BaseNode) GetID() string {
	return this.id
}

func (this *BaseNode) GetName() string {
	return this.name
}
func (this *BaseNode) GetTitle() string {
	//fmt.Println("GetTitle ", this.title)
	return this.title
}

// 这是将信号传播到该节点的主要方法. 此方法调用所有回调：“enter”、“open”、“tick”、“close”和“exit”.
// 它仅打开尚未打开的节点.同样，此方法仅在节点返回与"RUNNING"不同的状态时关闭该节点.
// 返回 Tick 状态
func (this *BaseNode) execute(tick *Tick) behavior.Status {
	//fmt.Println("_execute :", this.title)
	// ENTER
	this.enter(tick)

	// OPEN
	if !tick.Blackboard.GetBool("isOpen", tick.tree.id, this.id) {
		this.open(tick)
	}

	// TICK
	var status = this.tick(tick)

	// CLOSE
	if status != behavior.RUNNING {
		this.close(tick)
	}

	// EXIT
	this.exit(tick)

	return status
}

func (this *BaseNode) Execute(tick *Tick) behavior.Status {
	return this.execute(tick)
}

func (this *BaseNode) Stop(tick *Tick) {

}

// 输入方法的包装
func (this *BaseNode) enter(tick *Tick) {
	tick.enterNode(this)
	this.OnEnter(tick)
}

// 打开方法的包装
func (this *BaseNode) open(tick *Tick) {
	//fmt.Println("_open :", this.title)
	tick.openNode(this)
	tick.Blackboard.Set("isOpen", true, tick.tree.id, this.id)
	this.OnOpen(tick)
}

// Tick 方法的包装
func (this *BaseNode) tick(tick *Tick) behavior.Status {
	//fmt.Println("_tick :", this.title)
	tick.tickNode(this)
	return this.OnTick(tick)
}

// 关闭方法的包装
func (this *BaseNode) close(tick *Tick) {
	tick.closeNode(this)
	tick.Blackboard.Set("isOpen", false, tick.tree.id, this.id)
	this.OnClose(tick)
}

// 退出方法的包装
func (this *BaseNode) exit(tick *Tick) {
	tick.exitNode(this)
	this.OnExit(tick)
}
