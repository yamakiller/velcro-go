package trans

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime/debug"

	"github.com/yamakiller/velcro-go/rpc2/pkg/remote"
	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/rpc2/pkg/serviceinfo"
	"github.com/yamakiller/velcro-go/rpc2/pkg/stats"
	"github.com/yamakiller/velcro-go/utils/endpoint"
	"github.com/yamakiller/velcro-go/utils/verrors"
	"github.com/yamakiller/velcro-go/vlog"
)

func NewDefaultSvrTransHandler(opt *remote.ServerOption, addon ExtAddon) (remote.ServerTransHandler, error) {
	svrHdlr := &svrTransHandler{
		opt:           opt,
		codec:         opt.Codec,
		svcSearchMap:  opt.SvcMap,
		targetSvcInfo: opt.TargetSvcInfo,
		addon:         addon,
	}

	if svrHdlr.opt.TracerCtl == nil {
		// 如果跟踪控制器为空那么初始化跟踪控制器
		svrHdlr.opt.TracerCtl = &rpcinfo.TraceController{}
	}

	return svrHdlr, nil
}

type svrTransHandler struct {
	opt           *remote.ServerOption
	svcSearchMap  map[string]*serviceinfo.ServiceInfo
	targetSvcInfo *serviceinfo.ServiceInfo
	inkHdlFunc    endpoint.Endpoint
	codec         remote.Codec
	transPipe     *remote.TransPipeline
	addon         ExtAddon
}

// Write 实现远程服务处理器接口
func (t *svrTransHandler) Write(ctx context.Context, c net.Conn, sendMsg remote.Message) (nctx context.Context, err error) {
	var bufWriter remote.ByteBuffer
	ri := sendMsg.RPCInfo()
	rpcinfo.Record(ctx, ri, stats.WriteStart, nil)
	defer func() {
		// 1.释放缓冲区
		// 2.保存rpc写入记录
		t.addon.ReleaseBuffer(bufWriter, err)
		rpcinfo.Record(ctx, ri, stats.WriteFinish, err)
	}()

	svcInfo := sendMsg.ServiceInfo()
	if svcInfo != nil {
		if methodInfo, _ := GetMethodInfo(ri, svcInfo); methodInfo != nil {
			if methodInfo.OneWay() {
				return ctx, nil
			}
		}
	}

	bufWriter = t.addon.NewWriteByteBuffer(ctx, c, sendMsg)
	err = t.codec.Encode(ctx, sendMsg, bufWriter)
	if err != nil {
		return ctx, err
	}
	return ctx, bufWriter.Flush()
}

func (t *svrTransHandler) Read(ctx context.Context, conn net.Conn, recvMsg remote.Message) (nctx context.Context, err error) {
	var bufReader remote.ByteBuffer
	defer func() {
		// 1.释放缓冲区
		// 2.保存rpc读取记录
		t.addon.ReleaseBuffer(bufReader, err)
		rpcinfo.Record(ctx, recvMsg.RPCInfo(), stats.ReadFinish, err)
	}()
	rpcinfo.Record(ctx, recvMsg.RPCInfo(), stats.ReadStart, nil)

	bufReader = t.addon.NewReadByteBuffer(ctx, conn, recvMsg)
	if codec, ok := t.codec.(remote.MetaDecoder); ok {
		if err = codec.DecodeMeta(ctx, recvMsg, bufReader); err == nil {
			if t.opt.Profiler != nil && t.opt.ProfilerTransInfoTagging != nil && recvMsg.TransInfo() != nil {
				var tags []string
				ctx, tags = t.opt.ProfilerTransInfoTagging(ctx, recvMsg)
				ctx = t.opt.Profiler.Tag(ctx, tags...)
			}
			err = codec.DecodePayload(ctx, recvMsg, bufReader)
		}
	} else {
		err = t.codec.Decode(ctx, recvMsg, bufReader)
	}
	if err != nil {
		recvMsg.Tags()[remote.ReadFailed] = true
		return ctx, err
	}
	return ctx, nil
}

func (t *svrTransHandler) newCtxWithRPCInfo(ctx context.Context, conn net.Conn) (context.Context, rpcinfo.RPCInfo) {
	// 如果开启连接池
	if rpcinfo.PoolEnabled() {
		// 重用每个连接的 rpcinfo
		return ctx, rpcinfo.GetRPCInfo(ctx)
		// 延迟重新初始化以获得更快的响应
	}
	// 如果不开启连接池
	// 需要重新创建一个新的RPCINFO
	ri := t.opt.InitOrResetRPCInfoFunc(nil, conn.RemoteAddr())
	return rpcinfo.NewCtxWithRPCInfo(ctx, ri), ri
}

// OnRead 实现服务接口
// 返回错误后应关闭连接.
func (t *svrTransHandler) OnRead(ctx context.Context, c net.Conn) (err error) {
	ctx, ri := t.newCtxWithRPCInfo(ctx, c)
	t.addon.SetReadTimeout(ctx, c, ri.Config(), remote.Server)
	var recvMsg remote.Message
	var sendMsg remote.Message
	closeConnOutsideIfErr := true
	defer func() {
		panicErr := recover()
		var wrapErr error
		if panicErr != nil {
			stack := string(debug.Stack())
			if c != nil {
				ri := rpcinfo.GetRPCInfo(ctx)
				rService, rAddr := getRemoteInfo(ri, c)
				vlog.ContextErrorf(ctx, "VELCRO: panic happened, remoteAddress=%s, remoteService=%s, error=%v\nstack=%s", rAddr, rService, panicErr, stack)
			} else {
				vlog.ContextErrorf(ctx, "VELCRO: panic happened, error=%v\nstack=%s", panicErr, stack)
			}
			if err != nil {
				wrapErr = verrors.ErrPanic.WithCauseAndStack(fmt.Errorf("[happened in OnRead] %s, last error=%s", panicErr, err.Error()), stack)
			} else {
				wrapErr = verrors.ErrPanic.WithCauseAndStack(fmt.Errorf("[happened in OnRead] %s", panicErr), stack)
			}
		}
		t.finishTracer(ctx, ri, err, panicErr)
		t.finishProfiler(ctx)
		remote.RecycleMessage(recvMsg)
		remote.RecycleMessage(sendMsg)
		// 如果启动池化,则重置rpcinfo
		if rpcinfo.PoolEnabled() {
			t.opt.InitOrResetRPCInfoFunc(ri, c.RemoteAddr())
		}
		if wrapErr != nil {
			err = wrapErr
		}
		if err != nil && !closeConnOutsideIfErr {
			err = nil
		}
	}()

	ctx = t.startTracer(ctx, ri)
	ctx = t.startProfiler(ctx)
	recvMsg = remote.NewMessageWithNewer(t.targetSvcInfo,
		t.svcSearchMap,
		ri,
		remote.Call,
		remote.Server,
		t.opt.RefuseTrafficWithoutServiceName)
	recvMsg.SetPayloadCodec(t.opt.PayloadCodec)
	ctx, err = t.transPipe.Read(ctx, c, recvMsg)
	if err != nil {
		t.writeErrorReplyIfNeeded(ctx, recvMsg, c, err, ri, true)
		// t.OnError(ctx, err, conn) will be executed at outer function when transServer close the conn
		return err
	}

	svcInfo := recvMsg.ServiceInfo()
	// 关于心跳的处理
	// 如果指定的编解码器支持心跳, 则在之前的读取过程中, recvMsg.MessageType将被设置为remote.Heartbeat.
	if recvMsg.MessageType() == remote.Heartbeat {
		sendMsg = remote.NewMessage(nil, svcInfo, ri, remote.Heartbeat, remote.Server)
	} else {
		// 回复处理
		var methodInfo serviceinfo.MethodInfo
		if methodInfo, err = GetMethodInfo(ri, svcInfo); err != nil {
			// it won't be err, because the method has been checked in decode, err check here just do defensive inspection
			t.writeErrorReplyIfNeeded(ctx, recvMsg, c, err, ri, true)
			// for proxy case, need read actual remoteAddr, error print must exec after writeErrorReplyIfNeeded,
			// t.OnError(ctx, err, conn) will be executed at outer function when transServer close the conn
			return err
		}
		if methodInfo.OneWay() {
			sendMsg = remote.NewMessage(nil, svcInfo, ri, remote.Reply, remote.Server)
		} else {
			sendMsg = remote.NewMessage(methodInfo.NewResult(), svcInfo, ri, remote.Reply, remote.Server)
		}

		ctx, err = t.transPipe.OnMessage(ctx, recvMsg, sendMsg)
		if err != nil {
			// error 无法在此处包装打印, 因此必须在 NewTransError 之前执行.
			t.OnError(ctx, err, c)
			err = remote.NewTransError(remote.InternalError, err)
			if closeConn := t.writeErrorReplyIfNeeded(ctx, recvMsg, c, err, ri, false); closeConn {
				return err
			}
			// connection don't need to be closed when the error is return by the server handler
			closeConnOutsideIfErr = false
			return
		}
	}

	remote.FillSendMsgFromRecvMsg(recvMsg, sendMsg)
	if ctx, err = t.transPipe.Write(ctx, c, sendMsg); err != nil {
		return err
	}
	return
}

// OnMessage 实现remote.ServerTransHandler接口.
// 消息为解码后信息,既然 args, result
func (t *svrTransHandler) OnMessage(ctx context.Context, args, result remote.Message) (context.Context, error) {
	err := t.inkHdlFunc(ctx, args.Data(), result.Data())
	return ctx, err
}

// OnActive 实现remote.ServerTransHandler接口.
func (t *svrTransHandler) OnActive(ctx context.Context, c net.Conn) (context.Context, error) {
	rio := t.opt.InitOrResetRPCInfoFunc(nil, c.RemoteAddr())
	return rpcinfo.NewCtxWithRPCInfo(ctx, rio), nil
}

// OnInactive 实现remote.ServerTransHandler接口.
func (t *svrTransHandler) OnInactive(ctx context.Context, c net.Conn) {
	rpcinfo.PutRPCInfo((rpcinfo.GetRPCInfo(ctx)))
}

// OnError 实现remote.ServerTransHandler接口.
func (t *svrTransHandler) OnError(ctx context.Context, err error, c net.Conn) {
	rio := rpcinfo.GetRPCInfo(ctx)
	rService, rAddr := getRemoteInfo(rio, c)
	if t.addon.IsRemoteClosedErr(err) {
		if rio == nil {
			return
		}
		remote := rpcinfo.AsMutableEndpointInfo(rio.From())
		remote.SetTag(rpcinfo.RemoteClosedTag, "1")
	} else {
		var de *verrors.DetailedError
		if ok := errors.As(err, &de); ok && de.Stack() != "" {
			vlog.ContextErrorf(ctx, "VELCRO: rocessing request error, remoteService=%s, remoteAddr=%v, error=%s\nstack=%s",
				rService, rAddr, err.Error(), de.Stack())
			return
		}

		vlog.ContextErrorf(ctx, "VELCRO:: processing request error, remoteService=%s, remoteAddr=%v, error=%s",
			rService, rAddr, err.Error())
	}
}

// SetInvokeHandleFunc 实现 remote.InvokeHandleFuncSetter 接口.
func (t *svrTransHandler) SetInvokeHandleFunc(inkHdlFunc endpoint.Endpoint) {
	t.inkHdlFunc = inkHdlFunc
}

// SetPipeline 实现 remote.ServerTransHandler 接口.
func (t *svrTransHandler) SetPipeline(p *remote.TransPipeline) {
	t.transPipe = p
}

func (t *svrTransHandler) writeErrorReplyIfNeeded(ctx context.Context,
	recvMsg remote.Message,
	c net.Conn,
	err error,
	rio rpcinfo.RPCInfo,
	doOnMessage bool) (shouldCloseConn bool) {
	if cn, ok := c.(remote.IsActive); ok && !cn.IsActive() {
		// 连接已经关闭
		return
	}
	svcInfo := recvMsg.ServiceInfo()
	if svcInfo != nil {
		if methodInfo, _ := GetMethodInfo(rio, svcInfo); methodInfo != nil {
			if methodInfo.OneWay() {
				return
			}
		}
	}

	transErr, isTransErr := err.(*remote.TransError)
	if !isTransErr {
		return
	}
	errMsg := remote.NewMessage(transErr, svcInfo, rio, remote.Exception, remote.Server)
	remote.FillSendMsgFromRecvMsg(recvMsg, errMsg)
	if doOnMessage {
		// 如果在正常的 OnMessage 之前发生错误, 则执行它以将 header trans 信息传输到 rpcinfo.
		t.transPipe.OnMessage(ctx, recvMsg, errMsg)
	}
	ctx, err = t.transPipe.Write(ctx, c, errMsg)
	if err != nil {
		vlog.ContextErrorf(ctx, "VELCRO: write error reply failed, remote=%s, error=%s", c.RemoteAddr(), err.Error())
		return true
	}
	return
}

// startTracer 启动跟踪
func (t *svrTransHandler) startTracer(ctx context.Context, ri rpcinfo.RPCInfo) context.Context {
	c := t.opt.TracerCtl.DoStart(ctx, ri)
	return c
}

// finishTracer 完成跟中
func (t *svrTransHandler) finishTracer(ctx context.Context, ri rpcinfo.RPCInfo, err error, panicErr interface{}) {
	rpcStats := rpcinfo.AsMutableRPCStats(ri.Stats())
	if rpcStats == nil {
		return
	}
	if panicErr != nil {
		rpcStats.SetPanicked(panicErr)
	}
	if err != nil && t.addon.IsRemoteClosedErr(err) {
		// 不应将远程连接关闭引起的错误视为服务器错误.
		err = nil
	}
	t.opt.TracerCtl.DoFinish(ctx, ri, err)
	// 对于服务器端，rpcinfo 在连接上重用, 清除 rpc 统计信息但保留级别配置.
	sl := ri.Stats().Level()
	rpcStats.Reset()
	rpcStats.SetLevel(sl)
}

func (t *svrTransHandler) startProfiler(ctx context.Context) context.Context {
	if t.opt.Profiler == nil {
		return ctx
	}
	return t.opt.Profiler.Prepare(ctx)
}

func (t *svrTransHandler) finishProfiler(ctx context.Context) {
	if t.opt.Profiler == nil {
		return
	}
	t.opt.Profiler.Untag(ctx)
}

func getRemoteInfo(ri rpcinfo.RPCInfo, c net.Conn) (string, net.Addr) {
	rAddr := c.RemoteAddr()
	if ri == nil {
		return "", rAddr
	}
	if rAddr != nil && rAddr.Network() == "unix" {
		if ri.From().Address() != nil {
			rAddr = ri.From().Address()
		}
	}
	return ri.From().ServiceName(), rAddr
}
