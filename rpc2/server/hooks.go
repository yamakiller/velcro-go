package server

import "sync"

// RegisterStartHook 注册hook启动后函数
func RegisterStartHook(h func()) {
	muStartHooks.Lock()
	defer muStartHooks.Unlock()
	onServerStart.add(h)
}

// RegisterShutdownHook 注册hook 关闭函数
func RegisterShutdownHook(h func()) {
	muShutdownHooks.Lock()
	defer muShutdownHooks.Unlock()
	onShutdown.add(h)
}

// Hooks is a collection of func.
type Hooks []func()

// Add 添加hooks函数
func (h *Hooks) add(g func()) {
	*h = append(*h, g)
}

// Server hooks
var (
	onServerStart   Hooks      // 服务启动后执行
	muStartHooks    sync.Mutex // onServerStart 的同步锁
	onShutdown      Hooks      // 服务优雅关闭后执行
	muShutdownHooks sync.Mutex // onShutdown 的同步锁
)
