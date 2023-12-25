package proxy

import (
	"sync"
	"time"

	"github.com/yamakiller/velcro-go/cluster/repeat"
)

// RpcProxyConnRepeat 代理重试器
type RpcProxyConnRepeat struct {
	host         string
	conn         *RpcProxyConn
	dialTimeouot time.Duration
	printError   func(fmts string, args ...interface{})
	stopper      chan struct{}
	mu           sync.Mutex
	wg           sync.WaitGroup
}

func (rpcr *RpcProxyConnRepeat) start() {
	rpcr.stopper = make(chan struct{})
	rpcr.wg.Add(1)
	go func() {
		defer rpcr.wg.Done()
		repeat.Repeat(repeat.FnWithCounter(rpcr.reconnect),
			// 如果 reconnect 返回空停止流程
			repeat.StopOnSuccess(),
			repeat.WithDelay(repeat.ExponentialBackoff(500*time.Millisecond).Set()))

		rpcr.mu.Lock()
		if !rpcr.isStopped() {
			close(rpcr.stopper)
		}
		rpcr.mu.Unlock()
	}()
}

func (rpcr *RpcProxyConnRepeat) stop() {

	rpcr.mu.Lock()
	if !rpcr.isStopped() {
		close(rpcr.stopper)
	}
	rpcr.mu.Unlock()
	rpcr.wg.Wait()
	rpcr.conn = nil
	rpcr.stopper = nil
}

func (rpcr *RpcProxyConnRepeat) isStopped() bool {
	select {
	case <-rpcr.stopper:
		return true
	default:
		return false
	}
}

func (rpcr *RpcProxyConnRepeat) reconnect(counter int) error {
	if rpcr.isStopped() {
		return nil
	}

	if err := rpcr.conn.Dial(rpcr.host, rpcr.dialTimeouot); err != nil {
		rpcr.printError("%s connect fail[error:%s]", rpcr.host, err.Error())
		return err
	}

	rpcr.conn.proxy.Lock()
	rpcr.conn.proxy.alive[rpcr.host] = false
	rpcr.conn.proxy.Unlock()

	rpcr.conn.proxy.balancer.Add(rpcr.host)

	return nil
}
