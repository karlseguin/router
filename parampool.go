package router

import (
	"sync/atomic"
)

type ParamPool struct {
	misses int64
	size   int
	list   chan *Params
}

func NewParamPool(size, count int) *ParamPool {
	pool := &ParamPool{
		size: size,
		list: make(chan *Params, count),
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
