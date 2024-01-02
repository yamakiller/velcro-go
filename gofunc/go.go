package gofunc

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/yamakiller/velcro-go/profiler"
	"github.com/yamakiller/velcro-go/utils/gopool"
)

// GoTask is used to spawn a new task.
type GoTask func(context.Context, func())

// GoFunc is the default func used globally.
var GoFunc GoTask

func init() {
	GoFunc = func(ctx context.Context, f func()) {
		gopool.ContextGo(ctx, func() {
			profiler.Tag(ctx)
			f()
			profiler.Untag(ctx)
		})
	}
}

// RecoverGoFuncWithInfo 用于调试崩溃信息
func RecoverGoFuncWithInfo(ctx context.Context, task func(), info *Info) {
	GoFunc(ctx, func() {
		defer func() {
			if panicErr := recover(); panicErr != nil {
				stack := string(debug.Stack())
				if info.EscalateFailure != nil {
					info.EscalateFailure(fmt.Errorf("%s %v", info.Name, panicErr), stack)
				}
			}
		}()

		task()
	})
}

// Info is used to pass key info to go func, which is convenient for troubleshooting
type Info struct {
	Name            string
	EscalateFailure func(reason interface{}, message interface{})
}

// NewBasicInfo is to new Info with remoteService  and remoteAddr.
func NewBasicInfo(name string, failure func(reason interface{}, message interface{})) *Info {
	info := infoPool.Get().(*Info)
	info.Name = name
	info.EscalateFailure = failure
	return info
}

var (
	infoPool = &sync.Pool{
		New: func() interface{} {
			return new(Info)
		},
	}
)

/*func RecoverGoFuncWithInfo(ctx context.Context, task func(), info *Info) {
	GoFunc(ctx, func() {
		defer func() {
			if panicErr := recover(); panicErr != nil {
				if info.RemoteService == "" {
					info.RemoteService = "unknown"
				}
				if info.RemoteAddr == "" {
					info.RemoteAddr = "unknown"
				}
				stack := string(debug.Stack())
				vlog.ContextErrorf(ctx, "VELCRO: panic happened, remoteService=%s remoteAddress=%s error=%v\nstack=%s",
					info.RemoteService, info.RemoteAddr, panicErr, stack)

				phLock.RLock()
				if panicHandler != nil {
					panicHandler(info, panicErr, stack)
				}
				phLock.RUnlock()
			}
			infoPool.Put(info)
		}()

		task()
	})
}

func SetPanicHandler(hdlr func(info *Info, panicErr interface{}, panicStack string)) {
	phLock.Lock()
	panicHandler = hdlr
	phLock.Unlock()
}

// NewBasicInfo is to new Info with remoteService  and remoteAddr.
func NewBasicInfo(remoteService, remoteAddr string) *Info {
	info := infoPool.Get().(*Info)
	info.RemoteService = remoteService
	info.RemoteAddr = remoteAddr
	return info
}

// Info is used to pass key info to go func, which is convenient for troubleshooting
type Info struct {
	RemoteService string
	RemoteAddr    string
}

var (
	// EmptyInfo is empty Info which is convenient to reuse
	EmptyInfo = &Info{}

	panicHandler func(info *Info, panicErr interface{}, panicStack string)
	phLock       sync.RWMutex

	infoPool = &sync.Pool{
		New: func() interface{} {
			return new(Info)
		},
	}
)*/
