package streaming

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/yamakiller/velcro-go/utils/gopool"
	"github.com/yamakiller/velcro-go/utils/verrors"
)

func CallWithTimeout(timeout time.Duration, cancel context.CancelFunc, f func() (err error)) error {
	if timeout <= 0 {
		return f()
	}

	if cancel == nil {
		panic("nil cancel func. Please pass in the cancel func in the context (set in context BEFORE " +
			"the client method call, NOT just before recv/send), to avoid blocking recv/send calls")
	}

	begin := time.Now()
	timer := time.NewTimer(timeout)

	defer timer.Stop()
	finishChan := make(chan error, 1) // 非阻塞通道以避免 goroutine 泄漏
	gopool.Go(func() {
		var bizErr error
		defer func() {
			if r := recover(); r != nil {
				cancel()
				timeoutErr := fmt.Errorf("CallWithTimeout: panic in business code, timeout_config=%v, panic=%v, stack=%s",
					timeout, r, debug.Stack())
				bizErr = verrors.ErrRPCTimeout.WithCause(timeoutErr)
			}
			finishChan <- bizErr
		}()
		bizErr = f()
	})
	select {
	case <-timer.C:
		cancel()
		timeoutErr := fmt.Errorf("CallWithTimeout: timeout in business code, timeout_config=%v, actual=%v",
			timeout, time.Since(begin))
		return verrors.ErrRPCTimeout.WithCause(timeoutErr)
	case bizErr := <-finishChan:
		return bizErr
	}
}
