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
	return r.params.Get(key)
}
