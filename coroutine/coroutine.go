package coroutine

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type coState int32
type coFunc func([]interface{})

const (
	coIdle    = int32(0)
	coRun     = int32(1)
	coClosing = int32(2)
	coDeath   = int32(3)
)

var (
	// ErrPoolStoped Coroutine pool is closed
	ErrPoolStoped = errors.New("coroutine pool stoped")
)

type task struct {
	cb     coFunc
	params []interface{}
}

func (slf *coObject) run(cop *CoPools) {
	go func() {
		defer cop._wait.Done()
		for {
			for {
				select {
				case <-slf._q:
					slf._state = coDeath
					return
				case t := <-cop._taskQueue:
					atomic.CompareAndSwapInt32(&slf._state, coIdle, coRun)
					t.cb(t.params)
					t.params = nil
					atomic.CompareAndSwapInt32(&slf._state, coRun, coIdle)
				}
			}
		}
	}()
}

type coObject struct {
	_state int32
	_last  time.Duration
	_q     chan int
	_id    int
}

var (
	oneCoroutinePool     sync.Once
	defaultCoroutinePool *CoPools
)

//Instance doc
//@Summary coroutine pool instance
//@Method Instance
//@Return (*CoroutinePool)
func Instance() *CoPools {
	oneCoroutinePool.Do(func() {
		defaultCoroutinePool = &CoPools{}
	})
	return defaultCoroutinePool
}

//CoPools doc
//@Summary Coroutine pool
//@Struct CoPools
//@Member (int) number of task limit
type CoPools struct {
	_taskLimit int
	_maxNum    int
	_minNum    int

	_taskQueue chan task
	_runing    int32
	_curr      int32
	_seq       int32
	_cos       []coObject
	_wait      sync.WaitGroup
	_quit      chan int
}

func (slf *CoPools) scheduer() {
	defer slf._wait.Done()
	for {
		select {
		case <-slf._quit:
			goto scheduer_end
		default:
		}

		now := time.Now().Unix()
		for _, v := range slf._cos {
			if v._state == coRun ||
				v._state == coDeath ||
				v._state == coClosing {
				continue
			}

			if v._state == coIdle && ((time.Duration(now) - v._last) > time.Second*60) {
				if atomic.CompareAndSwapInt32(&v._state, int32(coIdle), int32(coClosing)) {
					close(v._q)
				}
			}
		}

		time.Sleep(time.Millisecond * 1000)
	}
scheduer_end:
	for i := 0; i < slf._maxNum; i++ {
		if slf._cos[i]._state != coDeath && slf._cos[i]._state != coClosing {
			atomic.StoreInt32(&slf._cos[i]._state, coClosing)
			close(slf._cos[i]._q)
		}
	}
}

//Start doc
//@Summary Start the coroutine pool
//@Method Start
//@Param (int) coroutine max of number
//@Param (int) coroutine min of number
//@Param (int) coroutine task max limit
func (slf *CoPools) Start(max int, min int, taskmax int) {
	slf._maxNum = max
	slf._minNum = min
	slf._taskLimit = taskmax

	if slf._maxNum == 0 {
		slf._maxNum = 65535
	}
	atomic.StoreInt32(&slf._seq, 1)

	now := time.Now().Unix()
	slf._taskQueue = make(chan task, slf._taskLimit)
	slf._cos = make([]coObject, slf._maxNum)
	for k, v := range slf._cos {
		v._id = k + 1
		v._state = coDeath
		v._last = time.Duration(now)
	}

	if slf._maxNum > 0 && slf._maxNum > slf._minNum {
		slf._wait.Add(1)
		go slf.scheduer()
	}

	for i := 0; i < slf._minNum; i++ {
		slf._cos[i]._state = coIdle
		slf.startOne(i)
	}
}

func (slf *CoPools) startOne(idx int) {
	slf._wait.Add(1)
	atomic.AddInt32(&slf._seq, 1)
	atomic.AddInt32(&slf._curr, 1)

	slf._cos[idx].run(slf)
}

//Stop doc
//@Summary Stop the coroutine pool
//@Method Stop
func (slf *CoPools) Stop() {
	close(slf._quit)
	slf._wait.Wait()
}

//Go doc
//@Summary Running coroutine
//@Method Go
//@Param  (func(params []interface{})) call function
//@Param  (...interface{}) call args
//@Return (error)
func (slf *CoPools) Go(f func(params []interface{}), params ...interface{}) error {

	select {
	case <-slf._quit:
		return ErrPoolStoped
	default:
	}
	runing := atomic.LoadInt32(&slf._runing)
	curr := atomic.LoadInt32(&slf._curr)

	if runing >= curr && curr < int32(slf._maxNum) {
		for i := 0; i < slf._maxNum; i++ {
			hash := ((int32(i) + slf._seq) % int32(slf._maxNum))
			if atomic.CompareAndSwapInt32(&slf._cos[hash]._state, coDeath, coIdle) {
				slf._seq = int32(hash) + slf._seq
				slf.startOne(int(hash))
				break
			}
		}
	}

	select {
	case <-slf._quit:
		return ErrPoolStoped
	case slf._taskQueue <- task{f, params}:
		return nil
	}
}
