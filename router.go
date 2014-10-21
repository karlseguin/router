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
			"HEAD":    newRoutePart(),
			"OPTIONS": newRoutePart(),
		},
	}
	router.paramPool = NewParamPool(config.paramPoolSize, config.paramPoolCount)
	return router
}

func (r *Router) NotFound(handler Handler) {
	r.notFound = handler
}

func (r *Router) All(path string, handler Handler) {
	for _, rp := range r.routes {
		r.add(rp, path, handler)
	}
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
	req := &Request{Request: httpReq, params: emptyParam}
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
	defer params.Release()

	var handler Handler
	for {
		original := rp
		index := strings.Index(path, "/")
		if index == -1 {
			index = len(path)
		}
		part := path[:index]
		if rp, ok = rp.parts[part]; ok == false {
			if original.prefixes != nil {
				lower := strings.ToLower(part)
				for _, prefix := range original.prefixes {
					if len(prefix.value) == 0 || strings.HasPrefix(lower, prefix.value) {
						rp = original
						handler = prefix.handler
						break
					}
				}
				if handler != nil {
					break
				}
			}
			if rp, ok = original.parts[":"]; ok == false {
				break
			}
			params.AddValue(part)
		}

		if rp == nil || len(path) == index {
			break
		}
		path = path[index+1:]
	}

	if rp == nil || (rp.handler == nil && handler == nil) {
		r.notFound(out, req)
		return
	}
	if l := params.Len(); l > 0 {
		for i := 0; i < l; i++ {
			params.SetKey(rp.params[i], i)
		}
		req.params = params
	}
	if handler == nil {
		handler = rp.handler
	}
	handler(out, req)

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

	params := make([]string, 0, 1)
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part[len(part)-1] == '*' {
			if rp.prefixes == nil {
				rp.prefixes = make([]*Prefix, 0, 1)
			}
			prefix := &Prefix{value: strings.ToLower(part[:len(part)-1]), handler: handler}
			rp.prefixes = appendPrefix(rp.prefixes, prefix)
			break
		}
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
	if rp.handler == nil {
		rp.handler = handler
	}
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

func appendPrefix(arr []*Prefix, value *Prefix) []*Prefix {
	target := arr
	if len(arr) == cap(arr) {
		target = make([]*Prefix, len(arr)+1)
		copy(target, arr)
	}
	return append(target, value)
}
