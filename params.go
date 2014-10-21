package router

var emptyParam = &Params{lookup: []struct{key, value string}{}}

type Params struct {
	lookup []struct{key, value string}
	length int
	pool   *ParamPool
}

func NewParams(pool *ParamPool, size int) *Params {
	return &Params{
		pool:   pool,
		lookup: make([]struct{key, value string}, size),
	}
}

func (p *Params) Len() int {
	return p.length
}

func (p *Params) AddValue(value string) {
	p.lookup[p.length] = struct{key, value string}{value: value}
	p.length += 1
}

func (p *Params) SetKey(key string, index int) {
	p.lookup[index].key = key
}

func (p *Params) Get(key string) string {
	for i, l := 0, p.length; i < l; i++ {
		if p.lookup[i].key == key {
			return p.lookup[i].value
		}
	}
	return ""
}

func (p *Params) Release() {
	if p.pool != nil {
		p.length = 0
		p.lookup = p.lookup
		p.pool.list <- p
	}
}
