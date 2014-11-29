package router

import (
	"github.com/karlseguin/params"
	"net/http"
)

type Request struct {
	*http.Request
	params params.Params
}

func (r *Request) Param(key string) string {
	value, _ := r.params.Get(key)
	return value
}
