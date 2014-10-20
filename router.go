package router

import (
	"net/http"
	"strings"
)

type Handler func(out http.ResponseWriter, req *Request)

type Router struct {
	notFound  Handler
	routes    map[string]*RoutePart
	paramPool *ParamPool
}

func New(config *Configuration) *Router {
	router := &Router{
		notFound: notFoundHandler,
		routes: map[string]*RoutePart{
			"GET":     newRoutePart(),
			"POST":    newRoutePart(),
			"PUT":     newRoutePart(),
			"DELETE":  newRoutePart(),
			"PURGE":   newRoutePart(),
			"PATCH":   newRoutePart(),
			"OPTIONS": newRoutePart(),
		},
	}
	router.paramPool = NewParamPool(config.paramPoolSize, config.paramPoolCount)
	return router
}

func (r *Router) NotFound(handler Handler) {
	r.notFound = handler
}

func (r *Router) Get(path string, handler Handler) {
	r.add(r.routes["GET"], path, handler)
}

func (r *Router) Post(path string, handler Handler) {
	r.add(r.routes["POST"], path, handler)
}

func (r *Router) Put(path string, handler Handler) {
	r.add(r.routes["PUT"], path, handler)
}

func (r *Router) Delete(path string, handler Handler) {
	r.add(r.routes["DELETE"], path, handler)
}

func (r *Router) Purge(path string, handler Handler) {
	r.add(r.routes["PURGE"], path, handler)
}

func (r *Router) Patch(path string, handler Handler) {
	r.add(r.routes["PATCH"], path, handler)
}

func (r *Router) Options(path string, handler Handler) {
	r.add(r.routes["OPTIONS"], path, handler)
}

func (r *Router) ServeHTTP(out http.ResponseWriter, httpReq *http.Request) {
	req := &Request{Request: httpReq}
	rp, ok := r.routes[req.Method]
	if ok == false {
		r.notFound(out, req)
		return
	}
	path := req.URL.Path
	if path == "" || path == "/" {
		handler := rp.handler
		if handler == nil {
			handler = r.notFound
		}
		handler(out, req)
		return
	}
	if path[0] == '/' {
		path = path[1:]
	}

	if path[len(path)-1] == '/' {
		path = path[:(len(path) - 1)]
	}
	params := r.paramPool.Checkout()

	for {
		original := rp
		index := strings.Index(path, "/")
		if index == -1 {
			index = len(path)
		}
		part := path[:index]

		if rp, ok = rp.parts[part]; ok == false {
			if rp, ok = original.parts[":"]; ok == false {
				break
			}
			params.values = append(params.values, part)
		}

		if len(path) == index {
			break
		}
		path = path[index+1:]
	}
	if rp == nil || rp.handler == nil {
		params.Release()
		r.notFound(out, req)
		return
	}
	l := len(params.values)
	if l > 0 {
		m := make(map[string]string, l)
		for i := 0; i < l; i++ {
			m[rp.params[i]] = params.values[i]
		}
		req.Params = m
	}
	params.Release()
	rp.handler(out, req)
}

func (r *Router) add(rp *RoutePart, path string, handler Handler) {
	if path == "" || path == "/" {
		rp.handler = handler
		return
	}

	if path[0] == '/' {
		path = path[1:]
	}
	if path[len(path)-1] == '/' {
		path = path[:(len(path) - 1)]
	}

	params := make([]string, 0, 2)
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part[0] == ':' {
			params = appendOne(params, part[1:])
			part = ":"
		}
		sub, exists := rp.parts[part]
		if exists == false {
			sub = newRoutePart()
			rp.parts[part] = sub
		}
		rp = sub
	}
	if len(params) > 0 {
		rp.params = params
	}
	rp.handler = handler
}

func (r Router) Routes() map[string]*RoutePart {
	return r.routes
}

func notFoundHandler(out http.ResponseWriter, req *Request) {
	out.WriteHeader(404)
}

func appendOne(arr []string, value string) []string {
	target := arr
	if len(arr) == cap(arr) {
		target = make([]string, len(arr)+1)
		copy(target, arr)
	}
	return append(target, value)
}
