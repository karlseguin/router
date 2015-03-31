package router

import (
	"gopkg.in/karlseguin/params.v2"
	"gopkg.in/karlseguin/scratch.v1"
	"net/http"
	"regexp"
	"strings"
)

type Action struct {
	Name    string
	Handler Handler
}

var (
	AllMethods  = []string{"GET", "POST", "PUT", "DELETE", "PURGE", "PATCH", "OPTIONS", "HEAD"}
	EmptyParams = params.New(0)
)

type Handler func(out http.ResponseWriter, req *Request)

type Router struct {
	notFound  *Action
	routes    map[string]*RoutePart
	ParamPool *params.Pool
	valuePool *scratch.StringsPool
}

func New(config *Configuration) *Router {
	router := &Router{
		routes:   make(map[string]*RoutePart),
		notFound: &Action{"", notFoundHandler},
	}
	router.ParamPool = params.NewPool(config.paramPoolSize, config.paramPoolCount)
	router.valuePool = scratch.NewStrings(config.paramPoolSize, config.paramPoolCount)
	return router
}

func (r *Router) NotFound(handler Handler) {
	r.notFound = &Action{"", handler}
}

func (r *Router) Add(method, path string, handler Handler) {
	r.AddNamed(method+":"+path, method, path, handler)
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

func (r *Router) Lookup(req *http.Request) (*params.Params, *Action) {
	rp, ok := r.routes[req.Method]
	params := EmptyParams
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
	var glob *RoutePart
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
					if strings.HasPrefix(lower, prefix.value) {
						rp = original
						action = prefix.action
						break
					}
				}
				if action != nil {
					break
				}
			}
			l := len(part)
			for _, param := range original.params {
				p := part
				if lp := len(param.suffix); lp > 0 {
					if strings.HasSuffix(p, param.suffix) == false {
						continue
					}
					p = part[:l-lp]
				}
				if param.constraint == nil || param.constraint.MatchString(p) {
					rp = param.route
					part = p
					break
				}
			}
			if rp == nil {
				break
			}
			values.Add(part)
		}

		if rp == nil || len(path) == index {
			break
		}
		if rp.glob {
			glob = rp
		}
		path = path[index+1:]
	}

	if rp == nil || rp.action == nil {
		if glob == nil {
			return params, nil
		}
		rp = glob
	}

	if rp.action == nil && action == nil {
		return params, nil
	}

	if l := values.Len(); l > 0 {
		params = r.ParamPool.Checkout()
		if lp := len(rp.variables); l > lp {
			l = lp
		}
		v := values.Values()
		for i := 0; i < l; i++ {
			params.Set(rp.variables[i], v[i])
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

	variables := make([]string, 0, 1)
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part[len(part)-1] == '*' {
			p := strings.ToLower(part[:len(part)-1])
			if len(p) == 0 {
				rp.glob = true
			} else {
				prefix := Prefix{value: p, action: action}
				rp.prefixes = append(rp.prefixes, prefix)
			}
			break
		}
		var sub *RoutePart
		if part[0] == ':' {
			var constraint *regexp.Regexp
			var suffix string
			variable := part[1:]
			if i := strings.IndexByte(variable, ':'); i != -1 {
				suffix = variable[i+1:]
				variable = variable[:i]
			}
			l := len(variable) - 1
			if variable[l] == ')' {
				if start := strings.IndexByte(variable, '('); start != -1 {
					constraint = regexp.MustCompile(variable[start+1 : l])
					variable = variable[:start]
				}
			}
			variables = append(variables, variable)
			for _, param := range rp.params {
				if param.constraint == nil && constraint == nil && len(param.suffix) == 0 && len(suffix) == 0 {
					sub = param.route
					break
				}
				if param.constraint == nil || constraint == nil || len(param.suffix) != 0 || len(suffix) != 0 {
					continue
				}
				if param.constraint.String() == constraint.String() && param.suffix == suffix {
					sub = param.route
					break
				}
			}

			if sub == nil {
				sub = newRoutePart()
				rp.params = append(rp.params, newParam(constraint, sub, suffix))
			}
		} else if sub = rp.parts[part]; sub == nil {
			sub = newRoutePart()
			rp.parts[part] = sub
		}
		rp = sub
	}
	if len(variables) > 0 {
		rp.variables = variables
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
