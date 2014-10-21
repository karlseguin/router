package router

var emptyParam = &Params{lookup: []string{}}

type Params struct {
	lookup []string
	length int
	pool   *ParamPool
}

func NewParams(pool *ParamPool, size int) *Params {
	return &Params{
		pool:   pool,
		lookup: make([]string, size),
	}
}

func (p *Params) Len() int {
	return p.length
}

func (p *Params) AddValue(value string) {
	p.lookup[p.length*2+1] = value
	p.length += 1
}

func (p *Params) AddKey(key string, index int) {
	p.lookup[index*2] = key
}

func (p *Params) Get(key string) string {
	for i, l := 0, p.length; i < l; i++ {
		position := i * 2
		if p.lookup[position] == key {
			return p.lookup[position+1]
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
