package syncx

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/yamakiller/velcro-go/utils"
)



func NewReentrantMutex() *ReentrantMutex {
	res := &ReentrantMutex{
		Mutex:     sync.Mutex{},
		recursion: 0,
		owner:     0,
	}
	return res
}

// ReentrantMutex 包装一个Mutex,实现可重入
type ReentrantMutex struct {
	sync.Mutex
	owner     int32 // 当前持有锁的goroutine id
	recursion int32 // 这个goroutine 重入的次数
}

func (m *ReentrantMutex) Lock() {
	fmt.Fprintln(os.Stderr,"ReentrantMutex11 ",m.owner, " ",utils.GetCurrentGoroutineID())
	if atomic.CompareAndSwapInt32(&m.owner, int32(utils.GetCurrentGoroutineID()), int32(utils.GetCurrentGoroutineID())) {
		fmt.Fprintln(os.Stderr,"ReentrantMutex ",m.owner, " ",utils.GetCurrentGoroutineID())
		m.recursion++
		return;
	}

	m.Mutex.Lock()
	
	atomic.StoreInt32(&m.owner, int32(utils.GetCurrentGoroutineID()))
	m.recursion = 1
}

func (m *ReentrantMutex) Unlock() {
	// 非持有锁的goroutine尝试释放锁，错误的使用
	if atomic.LoadInt32(&m.owner) ==int32(utils.GetCurrentGoroutineID())  {
		//panic(fmt.Sprintf("wrong the owner(%d): %d!", m.owner, gid))
		m.recursion--
		if m.recursion != 0 { // 如果这个goroutine还没有完全释放，则直接返回
			return
		}
	}

	// 此goroutine最后一次调用，需要释放锁
	atomic.StoreInt32(&m.owner, -1)
	m.Mutex.Unlock()
}