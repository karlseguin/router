package router

import (
	"net/http"
)

type Request struct {
	*http.Request
	params *Params
}

func (r *Request) Param(key string) string {
	return r.params.Get(key)
}
