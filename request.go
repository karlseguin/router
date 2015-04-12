package router

import (
	"gopkg.in/karlseguin/params.v2"
	"net/http"
	"net/url"
)

type Request struct {
	*http.Request
	query  url.Values
	params *params.Params
}

func (r *Request) Param(key string) string {
	value, _ := r.params.Get(key)
	return value
}

func (r *Request) Query(key string) string {
	return r.query.Get(key)
}
