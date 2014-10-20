package router

import (
	. "github.com/karlseguin/expect"
	"github.com/karlseguin/expect/build"
	"net/http"
	"net/http/httptest"
	"testing"
)

type RouterTests struct{}

func Test_Router(t *testing.T) {
	Expectify(new(RouterTests), t)
}

func (r *RouterTests) NotFound() {
	router := New(Configure())
	for _, path := range []string{"", "/", "it's", "/over/9000"} {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, build.Request().Path(path).Request)
		Expect(res.Code).To.Equal(404).Message("path: %s", path)
		Expect(res.Body.Len()).To.Equal(0)
	}
}

func (r *RouterTests) NotFoundWithCustomHandler() {
	router := New(Configure())
	router.NotFound(func(out http.ResponseWriter, req *Request) {
		out.WriteHeader(4004)
		out.Write([]byte("not found"))
	})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, build.Request().Request)
	Expect(res.Code).To.Equal(4004)
	Expect(res.Body.Bytes()).To.Equal([]byte("not found"))
}

func (r *RouterTests) DefaultRoute() {
	assertRouting("/", "/")
}

func (r *RouterTests) SimpleRoute() {
	assertRouting("/users", "/users")
	assertRouting("/users", "/users/")
	assertRouting("/users/", "/users")
	assertRouting("/users/", "/users/")
	assertNotFound("/users", "/user")
	assertNotFound("/users", "/users/323")
}

func (r *RouterTests) SimpleNestedRoute() {
	assertRouting("/users/all", "/users/all")
	assertNotFound("/users/all", "/users")
	assertNotFound("/users/all", "/users/")
	assertNotFound("/users/all", "/users/323/likes")
}

func (r *RouterTests) RouteWithParameter() {
	assertRouting("/users/:id", "/users/3233", "id", "3233")
	assertRouting("/users/:other_longer", "/users/ab & cd", "other_longer", "ab & cd")
	assertNotFound("/users/:id", "/users")
	assertNotFound("/users/:id", "/users/")
}

func (r *RouterTests) RouteWithParameterAndNesting() {
	assertRouting("/users/:id/likes", "/users/3233/likes", "id", "3233")
	// assertNotFound("/users/:id/likes", "/users/3233")
	// assertNotFound("/users/:id/likes", "/users/3233/like")
}

func (r *RouterTests) RouteWithMultipleParameter() {
	router := New(Configure())
	router.Get("/users/:id", testHandler("route-1"))
	router.Get("/users/:userId/likes", testHandler("route-2"))
	assertRouter(router, "/users/32", "route-1", "id", "32")
	assertRouter(router, "/users/32/likes", "route-2", "userId", "32")
}

func (r *RouterTests) RouteWithComplexSetup() {
	router := New(Configure())
	router.Get("/", testHandler("root"))
	router.Get("/users", testHandler("users"))
	router.Get("/users/:id", testHandler("users-id"))
	router.Get("/users/:userId/likes", testHandler("user-likes"))
	router.Get("/users/:userId/likes/:id", testHandler("user-likes-id"))

	assertRouter(router, "/", "root")
	assertRouter(router, "/users/", "users")
	assertRouter(router, "/users/944", "users-id", "id", "944")
	assertRouter(router, "/users/434/likes", "user-likes", "userId", "434")
	assertRouter(router, "/users/aaz/likes/4910a8", "user-likes-id", "userId", "aaz", "id", "4910a8")
}

func assertRouting(routePath, requestPath string, params ...string) {
	router := New(Configure())
	router.Get(routePath, func(out http.ResponseWriter, req *Request) {
		Expect(len(req.Params)).To.Equal(len(params) / 2)
		for i := 0; i < len(params); i += 2 {
			key := params[i]
			Expect(req.Params[key]).To.Equal(params[i+1])
		}
		out.WriteHeader(200)
		out.Write([]byte(routePath))
	})
	assertRouter(router, requestPath, routePath, params...)
}

func assertRouter(router *Router, requestPath string, body string, params ...string) {
	res := httptest.NewRecorder()
	req := build.Request().Path(requestPath).Request
	router.ServeHTTP(res, req)
	Expect(res.Code).To.Equal(200)
	Expect(string(res.Body.Bytes())).To.Equal(body)
}

func assertNotFound(routePath, requestPath string) {
	router := New(Configure())
	router.Get(routePath, func(out http.ResponseWriter, req *Request) {
		out.WriteHeader(200)
	})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, build.Request().Path(requestPath).Request)
	Expect(res.Code).To.Equal(404)
}

func testHandler(body string) func(out http.ResponseWriter, req *Request) {
	return func(out http.ResponseWriter, req *Request) {
		out.WriteHeader(200)
		out.Write([]byte(body))
	}
}
