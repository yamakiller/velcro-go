package retry

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yamakiller/velcro-go/gofunc"
	"github.com/yamakiller/velcro-go/rpc2/pkg/rpcinfo"
	"github.com/yamakiller/velcro-go/utils"
	"github.com/yamakiller/velcro-go/utils/circuitbreak"
	"github.com/yamakiller/velcro-go/utils/verrors"
	"github.com/yamakiller/velcro-go/vlog"
)

var errUnforeseen = errors.New("backup request: An unpredictable error occurred. Please record the error so that the code can be corrected.")

func newBackupRetryer(policy Policy, container *cbrContainer) (Retryer, error) {
	br := &backupRetryer{container: container}
	if err := br.UpdatePolicy(policy); err != nil {
		return nil, fmt.Errorf("newBackupRetryer failed, err=%w", err)
	}
	return br, nil
}

type backupRetryer struct {
	enable     bool
	policy     *BackupPolicy
	container  *cbrContainer
	retryDelay time.Duration
	sync.RWMutex
	errMsg string
}

type resultWrapper struct {
	rio rpcinfo.RPCInfo
	err error
}

// ShouldRetry implements the Retryer interface.
func (r *backupRetryer) ShouldRetry(ctx context.Context, err error, callTimes int, req interface{}, cbKey string) (string, bool) {
	r.RLock()
	defer r.RUnlock()
	if !r.enable {
		return "", false
	}
	if stop, msg := circuitBreakerStop(ctx, r.policy.StopPolicy, r.container, req, cbKey); stop {
		return msg, false
	}
	return "", true
}

// AllowRetry implements the Retryer interface.
func (r *backupRetryer) AllowRetry(ctx context.Context) (string, bool) {
	r.RLock()
	defer r.RUnlock()
	if !r.enable || r.policy.StopPolicy.MaxRetryTimes == 0 {
		return "", false
	}
	if stop, msg := chainStop(ctx, r.policy.StopPolicy); stop {
		return msg, false
	}
	return "", true
}

// Do implement the Retryer interface.
func (r *backupRetryer) Do(ctx context.Context, rpcCall RPCCallFunc, firstRI rpcinfo.RPCInfo, req interface{}) (lastRI rpcinfo.RPCInfo, recycleRI bool, err error) {
	r.RLock()
	retryTimes := r.policy.StopPolicy.MaxRetryTimes
	retryDelay := r.retryDelay
	r.RUnlock()
	var callTimes int32 = 0
	var callCosts utils.StringBuilder
	callCosts.RawStringBuilder().Grow(32)
	var recordCostDoing int32 = 0
	var abort int32 = 0
	finishedCount := 0
	// notice: buff num of chan is very important here, it cannot less than call times, or the below chan receive will block
	done := make(chan *resultWrapper, retryTimes+1)
	cbKey, _ := r.container.ctrl.GetKey(ctx, req)
	timer := time.NewTimer(retryDelay)
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			err = panicToErr(ctx, panicInfo, firstRI)
		}
		timer.Stop()
	}()
	// include first call, max loop is retryTimes + 1
	doCall := true
	for i := 0; ; {
		if doCall {
			doCall = false
			i++
			gofunc.GoFunc(ctx, func() {
				if atomic.LoadInt32(&abort) == 1 {
					return
				}
				var (
					e   error
					cRI rpcinfo.RPCInfo
				)
				defer func() {
					if panicInfo := recover(); panicInfo != nil {
						e = panicToErr(ctx, panicInfo, firstRI)
					}
					done <- &resultWrapper{cRI, e}
				}()
				ct := atomic.AddInt32(&callTimes, 1)
				callStart := time.Now()
				if r.container.enablePercentageLimit {
					// record stat before call since requests may be slow, making the limiter more accurate
					recordRetryStat(cbKey, r.container.panel, ct)
				}
				cRI, _, e = rpcCall(ctx, r)
				recordCost(ct, callStart, &recordCostDoing, &callCosts, &abort, e)
				if !r.container.enablePercentageLimit && r.container.stat {
					circuitbreak.RecordStat(ctx, req, nil, e, cbKey, r.container.ctrl, r.container.panel)
				}
			})
		}
		select {
		case <-timer.C:
			if _, ok := r.ShouldRetry(ctx, nil, i, req, cbKey); ok && i <= retryTimes {
				doCall = true
				timer.Reset(retryDelay)
			}
		case res := <-done:
			if res.err != nil && errors.Is(res.err, verrors.ErrRPCFinish) {
				// There will be only one request (goroutine) pass the `checkRPCState`, others will skip decoding
				// and return `ErrRPCFinish`, to avoid concurrent write to response and save the cost of decoding.
				// We can safely ignore this error and wait for the response of the passed goroutine.
				if finishedCount++; finishedCount >= retryTimes+1 {
					// But if all requests return this error, it must be a bug, preventive panic to avoid dead loop
					panic(errUnforeseen)
				}
				continue
			}
			atomic.StoreInt32(&abort, 1)
			recordRetryInfo(res.rio, atomic.LoadInt32(&callTimes), callCosts.String())
			return res.rio, false, res.err
		}
	}
}

// Prepare implements the Retryer interface.
func (r *backupRetryer) Prepare(ctx context.Context, prevRI, retryRI rpcinfo.RPCInfo) {
	handleRetryInstance(r.policy.RetrySameNode, prevRI, retryRI)
}

// UpdatePolicy implements the Retryer interface.
func (r *backupRetryer) UpdatePolicy(rp Policy) (err error) {
	if !rp.Enable {
		r.Lock()
		r.enable = rp.Enable
		r.Unlock()
		return nil
	}
	var errMsg string
	if rp.BackupPolicy == nil || rp.Type != BackupType {
		errMsg = "BackupPolicy is nil or retry type not match, cannot do update in backupRetryer"
		err = errors.New(errMsg)
	}
	if errMsg == "" && (rp.BackupPolicy.RetryDelayMillisecond == 0 || rp.BackupPolicy.StopPolicy.MaxRetryTimes < 0 ||
		rp.BackupPolicy.StopPolicy.MaxRetryTimes > maxBackupRetryTimes) {
		errMsg = "invalid backup request delay duration or retryTimes"
		err = errors.New(errMsg)
	}
	if errMsg == "" {
		if e := checkCircuitBreakerErrorRate(&rp.BackupPolicy.StopPolicy.CircuitBreakPolicy); e != nil {
			rp.BackupPolicy.StopPolicy.CircuitBreakPolicy.ErrorRate = defaultCircuitBreakerErrRate
			errMsg = fmt.Sprintf("backupRetryer %s, use default %0.2f", e.Error(), defaultCircuitBreakerErrRate)
			vlog.Warnf(errMsg)
		}
	}

	r.Lock()
	defer r.Unlock()
	r.enable = rp.Enable
	if err != nil {
		r.errMsg = errMsg
		return err
	}
	r.policy = rp.BackupPolicy
	r.retryDelay = time.Duration(rp.BackupPolicy.RetryDelayMillisecond) * time.Millisecond
	return nil
}

// AppendErrMsgIfNeeded implements the Retryer interface.
func (r *backupRetryer) AppendErrMsgIfNeeded(err error, ri rpcinfo.RPCInfo, msg string) {
	if verrors.IsTimeoutError(err) {
		// Add additional reason to the error message when timeout occurs but the backup request is not sent.
		appendErrMsg(err, msg)
	}
}

// Dump implements the Retryer interface.
func (r *backupRetryer) Dump() map[string]interface{} {
	r.RLock()
	defer r.RUnlock()
	if r.errMsg != "" {
		return map[string]interface{}{
			"enable":        r.enable,
			"backupRequest": r.policy,
			"errMsg":        r.errMsg,
		}
	}
	return map[string]interface{}{"enable": r.enable, "backupRequest": r.policy}
}

// Type implements the Retryer interface.
func (r *backupRetryer) Type() Type {
	return BackupType
}

// record request cost, it may execute concurrent
func recordCost(ct int32, start time.Time, recordCostDoing *int32, sb *utils.StringBuilder, abort *int32, err error) {
	if atomic.LoadInt32(abort) == 1 {
		return
	}
	for !atomic.CompareAndSwapInt32(recordCostDoing, 0, 1) {
		runtime.Gosched()
	}
	sb.WithLocked(func(b *strings.Builder) error {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(int(ct)))
		b.WriteByte('-')
		b.WriteString(strconv.FormatInt(time.Since(start).Microseconds(), 10))
		if err != nil && errors.Is(err, verrors.ErrRPCFinish) {
			// ErrRPCFinish means previous call returns first but is decoding.
			// Add ignore to distinguish.
			b.WriteString("(ignore)")
		}
		return nil
	})
	atomic.StoreInt32(recordCostDoing, 0)
}
