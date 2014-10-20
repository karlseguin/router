package router

import (
	"sync/atomic"
)

type ParamPool struct {
	misses   int64
	size int
	list     chan *Params
}

func NewParamPool(size, count int) *ParamPool {
	pool := &ParamPool{
		size: size,
		list:     make(chan *Params, count),
	}
	for i := 0; i < count; i++ {
		pool.list <- NewParams(pool, size)
	}
	return pool
}

func (p *ParamPool) Misses() int64 {
	return atomic.LoadInt64(&p.misses)
}

func (p *ParamPool) Checkout() *Params {
	select {
	case item := <-p.list:
		return item
	default:
		atomic.AddInt64(&p.misses, 1)
		return NewParams(nil, p.size)
	}
}

type Params struct {
	values []string
	pool *ParamPool
}

func NewParams(pool *ParamPool, size int) *Params {
	return &Params{
		pool: pool,
		values: make([]string, 0, size),
	}
}

func (p *Params) Release() {
	if p.pool != nil {
		p.values = p.values[0:0]
		p.pool.list <- p
	}
}
