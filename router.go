package router

import (
	"github.com/karlseguin/params"
	"github.com/karlseguin/scratch"
	"net/http"
	"strings"
)

type Action struct {
	Name string
	Handler Handler
}

var AllMethods = []string{"GET", "POST", "PUT", "DELETE", "PURGE", "PATCH", "OPTIONS", "HEAD" }

type Handler func(out http.ResponseWriter, req *Request)

type Router struct {
	notFound  *Action
	routes    map[string]*RoutePart
	paramPool *params.Pool
	valuePool *scratch.StringsPool
}

func New(config *Configuration) *Router {
	router := &Router{
		routes: make(map[string]*RoutePart),
		notFound: &Action{"", notFoundHandler},
	}
	router.paramPool = params.NewPool(config.paramPoolSize, config.paramPoolCount)
	router.valuePool = scratch.NewStrings(config.paramPoolSize, config.paramPoolCount)
	return router
}

func (r *Router) NotFound(handler Handler) {
	r.notFound = &Action{"", handler}
}

func (r *Router) Add(method, path string, handler Handler) {
	r.AddNamed(method + ":" + path, method, path, handler)
}

func (r *Router) AddNamed(name, method, path string, handler Handler) {
	if method == "ALL" {
		for _, m := range AllMethods {
			r.AddNamed(name, m, path, handler)
		}
		return
	}
	rp, exists := r.routes[method]
	if exists == false {
		rp = newRoutePart()
		r.routes[method] = rp
	}
	r.add(rp, path, &Action{name, handler})
}

func (r *Router) All(path string, handler Handler) {
	for _, method := range AllMethods {
		r.Add(method, path, handler)
	}
}

func (r *Router) AllNamed(name, path string, handler Handler) {
	for _, method := range AllMethods {
		r.AddNamed(name, method, path, handler)
	}
}

func (r *Router) Get(path string, handler Handler) {
	r.Add("GET", path, handler)
}

func (r *Router) Post(path string, handler Handler) {
	r.Add("POST", path, handler)
}

func (r *Router) Put(path string, handler Handler) {
	r.Add("PUT", path, handler)
}

func (r *Router) Delete(path string, handler Handler) {
	r.Add("DELETE", path, handler)
}

func (r *Router) Purge(path string, handler Handler) {
	r.Add("PURGE", path, handler)
}

func (r *Router) Patch(path string, handler Handler) {
	r.Add("PATCH", path, handler)
}

func (r *Router) Options(path string, handler Handler) {
	r.Add("OPTIONS", path, handler)
}

func (r *Router) ServeHTTP(out http.ResponseWriter, hr *http.Request) {
	params, action := r.Lookup(hr)
	defer params.Release()
	req := &Request{Request: hr, params: params}
	if action == nil || action.Handler == nil {
		r.notFound.Handler(out, req)
		return
	}
	action.Handler(out, req)
}

func (r *Router) Lookup(req *http.Request) (params.Params, *Action) {
	rp, ok := r.routes[req.Method]
	var params params.Params = params.Empty
	if ok == false {
		return params, nil
	}
	path := req.URL.Path
	if path == "" || path == "/" {
		action := rp.action
		if action == nil {
			action = r.notFound
		}
		return params, action
	}

	if path[0] == '/' {
		path = path[1:]
	}
	if path[len(path)-1] == '/' {
		path = path[:(len(path) - 1)]
	}

	values := r.valuePool.Checkout()
	defer values.Release()

	var action *Action
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
						action = prefix.action
						break
					}
				}
				if action != nil {
					break
				}
			}
			if rp, ok = original.parts[":"]; ok == false {
				break
			}
			values.Add(part)
		}

		if rp == nil || len(path) == index {
			break
		}
		path = path[index+1:]
	}

	if rp == nil || (rp.action == nil && action == nil) {
		return params, nil
	}

	if l := values.Len(); l > 0 {
		params = r.paramPool.Checkout()
		for i, value := range values.Values() {
			params.Set(rp.params[i], value)
		}
	}
	if action == nil {
		action = rp.action
	}
	return params, action
}

func (r *Router) add(rp *RoutePart, path string, action *Action) {
	if path == "" || path == "/" {
		rp.action = action
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
			prefix := &Prefix{value: strings.ToLower(part[:len(part)-1]), action: action}
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
	if rp.action == nil {
		rp.action = action
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
