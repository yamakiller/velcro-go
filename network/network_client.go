package network

import (
	"bufio"
	"io"
	"sync/atomic"

	"github.com/yamakiller/velcro-go/logs"
)

const (
	stateAccept int32 = iota
	stateAlive
	stateClosing
	stateClosed
)

/*func newClientContext(system *NetworkSystem, conn net.Conn) error {

	ids := system._handlers.NextId()
	cid, err := system._props._spawnHandler(system, ids)
	if err != nil {
		return err
	}

	this := &clientContext{
		_self:   cid,
		_system: system,
		_closer: conn,
		_reader: bufio.NewReaderSize(conn.(io.ReadWriteCloser), 1024),
		_writer: bufio.NewWriterSize(conn.(io.ReadWriteCloser), 1024),
	}

	// 翻入到ID集里

	return nil
}*/

type Context interface {
	// Client
	Client() Client
	// Self clientId
	Self() *ClientId
	// NetworkSystem System object
	NetworkSystem() *NetworkSystem
	// Reader 读取流
	Reader() *bufio.Reader
	// Writer 写入流
	Writer() *bufio.Writer
	// Logger 日志接口
	Logger() logs.LogAgent
}

type clientContext struct {
	_client Client
	_system *NetworkSystem
	_closer io.Closer
	_reader *bufio.Reader
	_writer *bufio.Writer
	_self   *ClientId
	_state  int32
}

func (ctx *clientContext) NetworkSystem() *NetworkSystem {
	return ctx._system
}

func (ctx *clientContext) Logger() logs.LogAgent {
	return ctx._system.Logger()
}

func (ctx *clientContext) Self() *ClientId {
	return ctx._self
}

func (ctx *clientContext) Client() Client {
	return ctx._client
}

func (ctx *clientContext) Reader() *bufio.Reader {
	return ctx._reader
}

func (ctx *clientContext) Writer() *bufio.Writer {
	return ctx._writer
}

func (ctx *clientContext) postUsrMessage(cid *ClientId, message interface{}) {
	cid.PostUsrMessage(ctx._system, message)
}

// Interface: MessageInvoker
func (ctx *clientContext) invokeSysMessage(message interface{}) {
	switch msg := message.(type) {
	case *Activation:
		ctx.onActivation()
	case *Close:
		ctx.onClose(msg)
	default:
		ctx._system.Logger().Error("CONTEXT", "System message unfound:%+v", msg)
	}
}

func (ctx *clientContext) invokeUsrMessage(md interface{}) {
	if atomic.LoadInt32(&ctx._state) == stateClosed {
		// already closed
		return
	}

	ctx.processMessage(md)
}

func (ctx *clientContext) onClose(message interface{}) {
	if atomic.LoadInt32(&ctx._state) >= stateClosing {
		// already closing
		return
	}

	atomic.StoreInt32(&ctx._state, stateClosing)

	ctx._closer.Close()
	ctx.invokeUsrMessage(closingMessage)
	ctx.finalizeStop()
}

func (ctx *clientContext) onActivation() {
	if atomic.LoadInt32(&ctx._state) != stateAccept {
		return
	}

	atomic.StoreInt32(&ctx._state, stateAlive)
}

func (ctx *clientContext) finalizeStop() {
	ctx._system._handlers.Remove(ctx._self)
	ctx.invokeUsrMessage(closedMessage)

	atomic.StoreInt32(&ctx._state, stateClosed)
}

func (ctx *clientContext) processMessage(m interface{}) {

	switch msg := m.(type) {
	case *Closing:
		ctx.Client().Closing(ctx._self)
	case *Closed:
		ctx.Client().Closed(ctx._self)
	default:
		ctx.Client().Send(ctx, msg)
	}
}

func (ctx *clientContext) incarnateClient() {
	atomic.StoreInt32(&ctx._state, stateAccept)

	/*ctx.actor = ctx.props.producer(ctx.actorSystem)

	metricsSystem, ok := ctx.actorSystem.Extensions.Get(extensionId).(*Metrics)
	if ok && metricsSystem.enabled {
		_ctx := context.Background()
		if instruments := metricsSystem.metrics.Get(metrics.InternalActorMetrics); instruments != nil {
			instruments.ActorSpawnCount.Add(_ctx, 1, metric.WithAttributes(metricsSystem.CommonLabels(ctx)...))
		}
	}*/
}
