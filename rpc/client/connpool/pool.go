package clientpool

import (
	"errors"
	"sync"
	"time"
)

func NewPool(new func(time.Duration) (any, error), close func(interface{}) error, maxIdle int,maxIdleTimeout time.Duration) *Pool {
	return &Pool{
		new:            new,
		close:          close,
		openingConns:   0,
		maxIdle:        maxIdle,
		maxIdleTimeout: maxIdleTimeout,
		pls:            make(chan interface{}, maxIdle),
	}
}

type Pool struct {
	pls            chan interface{}
	new            func(time.Duration) (any, error)
	close          func(interface{}) error
	openingConns   int
	maxIdle        int
	maxIdleTimeout time.Duration
	mu             sync.Mutex
}

func (p *Pool) Get() (interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	select {
	case conn := <-p.pls:
		return conn, nil
	default:
		if p.openingConns >= p.maxIdle {
			return nil, errors.New("pool is exhausted")
		}

		ret, err := p.new(p.maxIdleTimeout)
		if err != nil {
			return nil, err
		}
		p.openingConns++
		return ret, nil
	}
}

func (p *Pool) Put(conn interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	select {
	case p.pls <- conn:
		return nil
	default:
		p.close(conn)
	}
	return nil
}

func (p *Pool) Len() int{
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.pls)
}

func (p *Pool) Close(conn interface{}) error {
	if conn == nil {
		return errors.New("connection is nil")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.close == nil {
		return nil
	}
	p.openingConns--
	return p.close(conn)
}
