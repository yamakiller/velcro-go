package streaming

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/yamakiller/velcro-go/rpc/utils/verrors"
	"github.com/yamakiller/velcro-go/utils/gopool"
)

// CallWithTimeout 执行一个有超时的函数.
// 如果timeout为 0,函数将不超时执行;在这种情况下恐慌没有恢复;
// 如果时间用完,则会返回 verrors.ErrRPCTimeout;
// 如果你的函数发生恐慌, 它还会返回一个 kerrors.ErrRPCTimeout 以及恐慌详细信息;
// 其他类型的错误总是由您的函数返回.
//
// NOTE: “cancel”函数对于取消底层传输是必要的, 否则recv/send调用将阻塞直到传输关闭, 这可能会导致goroutine激增.
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
	finishChan := make(chan error, 1) // non-blocking channel to avoid goroutine leak

	gopool.Go(func() {
		var bizErr error
		defer func() {
			if r := recover(); r != nil {
				cancel()
				timeoutErr := fmt.Errorf(
					"CallWithTimeout: panic in business code, timeout_config=%v, panic=%v, stack=%s",
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
