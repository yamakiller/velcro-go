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

func NewNode(mutex sync.Locker) *Node {
	var curTime = time.Now()
	epoch := time.Now().UnixNano()/1e6 - 86400000

	newNode := &Node{
		_stepMask:  -1 ^ (-1 << stepBits),
		_timeShift: stepBits,
		_epoch:     curTime.Add(time.Unix(epoch/1000, (epoch%1000)*1000000).Sub(curTime)),
		_mu:        mutex,
	}

	return newNode
}

// Node 产生的ID 全局不唯一只能保障本节点唯一
type Node struct {
	_epoch     time.Time
	_time      int64
	_step      int64
	_stepMask  int64
	_timeShift uint8
	_mu        sync.Locker
}

func (n *Node) Generate() ID {
	n._mu.Lock()
	defer n._mu.Unlock()

	now := time.Since(n._epoch).Milliseconds()

	if now == n._time {
		n._step = (n._step + 1) & n._stepMask

		if n._step == 0 {
			for now <= n._time {
				now = time.Since(n._epoch).Milliseconds()
			}
		}
	} else {
		n._step = 0
	}

	n._time = now
	r := ID(((now) << n._timeShift) | (n._step))

	return r
}
