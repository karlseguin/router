package router

import (
	"testing"
	. "github.com/karlseguin/expect"
)

type ParamsTests struct {}

func Test_Params(t *testing.T) {
	Expectify(new(ParamsTests), t)
}

func (pt *ParamsTests) GetReturnsEmptyWhenNoParams() {
	p := &Params{}
	Expect(p.Len()).To.Equal(0)
	Expect(p.Get("spice")).To.Equal("")
}

func (pt *ParamsTests) BuildsAKeyValue() {
	p := NewParams(nil, 10)
	p.AddValue("value-1")
	p.AddValue("value-2")
	p.AddValue("value-3")
	p.AddKey("key-2", 1)
	p.AddKey("key-1", 0)
	p.AddKey("key-3", 2)
	Expect(p.Len()).To.Equal(3)
	Expect(p.Get("key-3")).To.Equal("value-3")
	Expect(p.Get("key-2")).To.Equal("value-2")
	Expect(p.Get("key-1")).To.Equal("value-1")
}
