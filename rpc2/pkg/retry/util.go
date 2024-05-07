package retry

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"

	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo/remoteinfo"
	"github.com/yamakiller/velcro-go/utils/metainfo"
	"github.com/yamakiller/velcro-go/utils/verrors"
	"github.com/yamakiller/velcro-go/vlog"
)

type ctxKey string

const (
	// 当 req 是重试调用时 TransitKey 会持久传输
	TransitKey = "RetryReq"
	// 用于忽略RPC请求并发写入
	ContextRequestOp ctxKey = "V_REQ_OP"
	// ContextRespOp 用于忽略 RPC 响应并发写入/读取
	ContextRespOp ctxKey = "V_RESP_OP"
	// 任何方法
	Wildcard = "*"
)

// Req或Resp操作状态，仅在可能发生并发写入时有用
const (
	OpNo int32 = iota
	OpDoing
	OpDone
)

var tagValueFirstTry = "0"

// DDLStopFunc is the definition of ddlStop func
type DDLStopFunc func(ctx context.Context, policy StopPolicy) (bool, string)

var ddlStopFunc DDLStopFunc

// RegisterDDLStop registers ddl stop.
func RegisterDDLStop(f DDLStopFunc) {
	ddlStopFunc = f
}

// If Ingress is turned on in the current node, check whether RPC_PERSIST_DDL_REMAIN_TIME exists,
// if it exists calculate handler time consumed by RPC_PERSIST_INGRESS_START_TIME and current time,
// if handler cost > ddl remain time, then do not execute retry.
func ddlStop(ctx context.Context, policy StopPolicy) (bool, string) {
	if !policy.DDLStop {
		return false, ""
	}
	if ddlStopFunc == nil {
		vlog.Warnf("enable ddl stop for retry, but no ddlStopFunc is registered")
		return false, ""
	}
	return ddlStopFunc(ctx, policy)
}

func chainStop(ctx context.Context, policy StopPolicy) (bool, string) {
	if policy.DisableChainStop {
		return false, ""
	}
	if !IsRemoteRetryRequest(ctx) {
		return false, ""
	}
	return true, "chain stop retry"
}

func circuitBreakerStop(ctx context.Context, policy StopPolicy, container *cbrContainer, request interface{}, cbrKey string) (bool, string) {
	if container.ctrl == nil || container.panel == nil {
		return false, ""
	}
	metricer := container.panel.GetMetricer(cbrKey)
	errRate := metricer.ErrorRate()
	sample := metricer.Samples()
	if sample < circuitBreakerMinSample || errRate < policy.CircuitBreakPolicy.ErrorRate {
		return false, ""
	}
	return true, fmt.Sprintf("retry circuit break, errRate=%0.3f, sample=%d", errRate, sample)
}

func handleRetryInstance(retrySameNode bool, prevRI, retryRI rpcinfo.RPCInfo) {
	calledInst := remoteinfo.AsRemoteInfo(prevRI.To()).GetInstance()
	if calledInst == nil {
		return
	}
	if retrySameNode {
		remoteinfo.AsRemoteInfo(retryRI.To()).SetInstance(calledInst)
	} else {
		if me := remoteinfo.AsRemoteInfo(retryRI.To()); me != nil {
			me.SetTag(rpcinfo.RetryPrevInstTag, calledInst.Address().String())
		}
	}
}

func makeRetryErr(ctx context.Context, msg string, callTimes int32) error {
	var ctxErr string
	if ctx.Err() == context.Canceled {
		ctxErr = "context canceled by business."
	}

	ri := rpcinfo.GetRPCInfo(ctx)
	to := ri.To()

	errMsg := fmt.Sprintf("retry[%d] failed, %s, to=%s, method=%s", callTimes-1, msg, to.ServiceName(), to.Method())
	target := to.Address()
	if target != nil {
		errMsg = fmt.Sprintf("%s, remote=%s", errMsg, target.String())
	}
	if ctxErr != "" {
		errMsg = fmt.Sprintf("%s, %s", errMsg, ctxErr)
	}
	return verrors.ErrRetry.WithCause(errors.New(errMsg))
}

func panicToErr(ctx context.Context, panicInfo interface{}, ri rpcinfo.RPCInfo) error {
	toService, toMethod := "unknown", "unknown"
	if ri != nil {
		toService, toMethod = ri.To().ServiceName(), ri.To().Method()
	}
	err := fmt.Errorf("KITEX: panic in retry, to_service=%s to_method=%s error=%v\nstack=%s",
		toService, toMethod, panicInfo, debug.Stack())
	vlog.ContextErrorf(ctx, "%s", err.Error())
	return err
}

func appendErrMsg(err error, msg string) {
	if e, ok := err.(*verrors.DetailedError); ok {
		// append no retry reason
		e.WithExtraMsg(msg)
	}
}

func recordRetryInfo(ri rpcinfo.RPCInfo, callTimes int32, lastCosts string) {
	if callTimes > 1 {
		if re := remoteinfo.AsRemoteInfo(ri.To()); re != nil {
			re.SetTag(rpcinfo.RetryTag, strconv.Itoa(int(callTimes)-1))
			// record last cost
			re.SetTag(rpcinfo.RetryLastCostTag, lastCosts)
		}
	}
}

// IsLocalRetryRequest checks whether it's a retry request by checking the RetryTag set in rpcinfo
// It's supposed to be used in client middlewares
func IsLocalRetryRequest(ctx context.Context) bool {
	ri := rpcinfo.GetRPCInfo(ctx)
	retryCountStr := ri.To().DefaultTag(rpcinfo.RetryTag, tagValueFirstTry)
	return retryCountStr != tagValueFirstTry
}

// IsRemoteRetryRequest checks whether it's a retry request by checking the TransitKey in metainfo
// It's supposed to be used in server side (handler/middleware)
func IsRemoteRetryRequest(ctx context.Context) bool {
	_, isRetry := metainfo.GetPersistentValue(ctx, TransitKey)
	return isRetry
}
