package snowflakealien

import (
	"sync"
	"time"
)

type ID int64

const (
	// 每毫秒里的最大计数值
	stepBits uint8 = 22
)

var (
	_mutex     sync.Mutex
	_once      sync.Once
	_epoch     time.Time
	_time      int64
	_step      int64
	_stepMask  int64
	_timeShift uint8
)

func Generate() ID {
	_once.Do(func() {
		_stepMask = -1 ^ (-1 << stepBits)
		_timeShift = stepBits

		var curTime = time.Now()
		epoch := time.Now().UnixNano()/1e6 - 86400000
		_epoch = curTime.Add(time.Unix(epoch/1000, (epoch%1000)*1000000).Sub(curTime))
	})

	_mutex.Lock()
	defer _mutex.Unlock()

	now := time.Since(_epoch).Milliseconds()

	if now == _time {
		_step = (_step + 1) & _stepMask

		if _step == 0 {
			for now <= _time {
				now = time.Since(_epoch).Milliseconds()
			}
		}
	} else {
		_step = 0
	}

	_time = now
	r := ID(((now) << _timeShift) | (_step))

	return r
}
