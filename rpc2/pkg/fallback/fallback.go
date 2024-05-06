package fallback

import (
	"context"
	"reflect"

	"github.com/yamakiller/velcro-go/rpc2/pkg/interfaces"
	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/vlog"
)

// ErrorFallback is to build fallback policy for error.
func ErrorFallback(ef Func) *Policy {
	return &Policy{fallbackFunc: func(ctx context.Context, req interfaces.VelcroArgs, resp interfaces.VelcroResult, err error) (fbErr error) {
		if err == nil {
			return err
		}
		return ef(ctx, req, resp, err)
	}}
}

// Policy is the definition for fallback.
//   - fallbackFunc is fallback func.
//   - reportAsFallback is used to decide whether to report Metric according to the Fallback result.
type Policy struct {
	fallbackFunc     Func
	reportAsFallback bool
}

func (p *Policy) EnableReportAsFallback() *Policy {
	p.reportAsFallback = true
	return p
}

// IsPolicyValid 策略是否有效.
func IsPolicyValid(p *Policy) bool {
	return p != nil && p.fallbackFunc != nil
}

// UnwrapHelper 帮助获取真实的请求和响应.
// 因此，RealReqRespFunc只需要处理真实的rpc req和resp, 而不需要处理XXXArgs和XXXResult.
// 例如: `client.WithFallback(fallback.NewFallbackPolicy(fallback.UnwrapHelper(yourRealReqRespFunc)))`
func UnwrapHelper(userFB RealReqRespFunc) Func {
	return func(ctx context.Context, args interfaces.VelcroArgs, result interfaces.VelcroResult, err error) error {
		req, resp := args.GetFirstArgument(), result.GetResult()
		fbResp, fbErr := userFB(ctx, req, resp, err)
		if fbResp != nil {
			result.SetSuccess(fbResp)
		}
		return fbErr
	}
}

// Func 是 fallback func的定义,它可以对error和resp进行fallback.
// 注意  args和result不是真正的rpc req和resp， 分别是生成代码的XXXArgs和XXXResult。
// 例如: client.WithFallback(fallback.NewFallbackPolicy(yourFunc))
type Func func(ctx context.Context, args interfaces.VelcroArgs, result interfaces.VelcroResult, err error) (fbErr error)

// RealReqRespFunc 是以真实 rpc req 作为参数的后备函数的定义, 并且必须返回真实的 rpc resp.
type RealReqRespFunc func(ctx context.Context, req, resp interface{}, err error) (fbResp interface{}, fbErr error)

// DoIfNeeded do fallback
func (p *Policy) DoIfNeeded(ctx context.Context, ri rpcinfo.RPCInfo, args, result interface{}, err error) (fbErr error, reportAsFallback bool) {
	if p == nil {
		return err, false
	}
	ka, kaOK := args.(interfaces.VelcroArgs)
	kr, krOK := result.(interfaces.VelcroResult)
	if !kaOK || !krOK {
		vlog.Warn("VELCRO: fallback cannot be supported, the args and result must be VelcroArgs and VelcroResult")
		return err, false
	}
	err, allowReportAsFB := getBizErrIfExist(ri, err)
	reportAsFallback = allowReportAsFB && p.reportAsFallback

	fbErr = p.fallbackFunc(ctx, ka, kr, err)

	if fbErr == nil && reflect.ValueOf(kr.GetResult()).IsNil() {
		vlog.Warnf("VELCRO: both fallback resp and error are nil, return original err=%v, to_service=%s to_method=%s",
			err, ri.To().ServiceName(), ri.To().Method())
		return err, false
	}
	return fbErr, reportAsFallback
}

func getBizErrIfExist(ri rpcinfo.RPCInfo, err error) (error, bool) {
	if err == nil {
		if bizErr := ri.Invocation().BizStatusErr(); bizErr != nil {
			err = bizErr
		}

		return err, false
	}
	return err, true
}
