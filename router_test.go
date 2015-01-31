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

func (_ RouterTests) NotFound() {
	router := New(Configure())
	for _, path := range []string{"", "/", "it's", "/over/9000"} {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, build.Request().Path(path).Request)
		Expect(res.Code).To.Equal(404).Message("path: %s", path)
		Expect(res.Body.Len()).To.Equal(0)
	}
}

func (_ RouterTests) NotFoundWithCustomHandler() {
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

func (_ RouterTests) DefaultRoute() {
	assertRouting("/", "/")
}

func (_ RouterTests) RouteWithNoParams() {
	var called bool
	router := New(Configure())
	router.Delete("/harkonen", func(out http.ResponseWriter, req *Request) {
		Expect(req.Param("friends")).To.Equal("")
		called = true
	})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, build.Request().Method("DELETE").Path("/harkonen").Request)
	Expect(called).To.Equal(true)
}

func (_ RouterTests) RouterToAll() {
	router := New(Configure())
	router.All("/power", testHandler("9000"))
	assertRouter(router, "GET", "/power", "9000")
	assertRouter(router, "POST", "/power", "9000")
	assertRouter(router, "PUT", "/power", "9000")
	assertRouter(router, "DELETE", "/power", "9000")
	assertRouter(router, "PATCH", "/power", "9000")
	assertRouter(router, "PURGE", "/power", "9000")
	assertRouter(router, "HEAD", "/power", "9000")
	assertRouter(router, "OPTIONS", "/power", "9000")
}

func (_ RouterTests) RouterToAllOverwrite() {
	router := New(Configure())
	router.Get("/power", testHandler("get-9000"))
	router.All("/power", testHandler("9000"))
	assertRouter(router, "GET", "/power", "get-9000")
	assertRouter(router, "POST", "/power", "9000")
	assertRouter(router, "PUT", "/power", "9000")
	assertRouter(router, "DELETE", "/power", "9000")
	assertRouter(router, "PATCH", "/power", "9000")
	assertRouter(router, "PURGE", "/power", "9000")
	assertRouter(router, "HEAD", "/power", "9000")
	assertRouter(router, "OPTIONS", "/power", "9000")
}

func (_ RouterTests) SimpleRoute() {
	assertRouting("/users", "/users")
	assertRouting("/users", "/users/")
	assertRouting("/users/", "/users")
	assertRouting("/users/", "/users/")
	assertNotFound("/users", "GET", "/user")
	assertNotFound("/users", "GET", "/users/323")
}

func (_ RouterTests) SimpleNestedRoute() {
	assertRouting("/users/all", "/users/all")
	assertNotFound("/users/all", "GET", "/users")
	assertNotFound("/users/all", "GET", "/users/")
	assertNotFound("/users/all", "GET", "/users/323/likes")
}

func (_ RouterTests) RouteWithParameter() {
	assertRouting("/users/:id", "/users/3233", "id", "3233")
	assertRouting("/users/:other_longer", "/users/ab & cd", "other_longer", "ab & cd")
	assertNotFound("/users/:id", "GET", "/users")
	assertNotFound("/users/:id", "GET", "/users/")
}

func (_ RouterTests) RouteWithParameterAndNesting() {
	assertRouting("/users/:id/likes", "/users/3233/likes", "id", "3233")
	assertNotFound("/users/:id/likes", "GET", "/users/3233")
	assertNotFound("/users/:id/likes", "GET", "/users/3233/like")
}

func (_ RouterTests) RouteWithMultipleParameter() {
	router := New(Configure())
	router.Get("/users/:id", testHandler("route-1"))
	router.Get("/users/:userId/likes", testHandler("route-2"))
	assertRouter(router, "GET", "/users/32", "route-1", "id", "32")
	assertRouter(router, "GET", "/users/32/likes", "route-2", "userId", "32")
}

func (_ RouterTests) RouteWithComplexSetup() {
	router := New(Configure())
	router.Get("/", testHandler("root"))
	router.Get("/users", testHandler("users"))
	router.Get("/users/:id", testHandler("users-id"))
	router.Get("/users/:userId/likes", testHandler("user-likes"))
	router.Get("/users/:userId/likes/:id", testHandler("user-likes-id"))

	assertRouter(router, "GET", "/", "root")
	assertRouter(router, "GET", "/users/", "users")
	assertRouter(router, "GET", "/users/944", "users-id", "id", "944")
	assertRouter(router, "GET", "/users/434/likes", "user-likes", "userId", "434")
	assertRouter(router, "GET", "/users/aaz/likes/4910a8", "user-likes-id", "userId", "aaz", "id", "4910a8")
}

func (_ RouterTests) RoutingWithPrefix() {
	router := New(Configure())
	router.Put("/", testHandler("root"))
	router.Put("/admin/*", testHandler("admin-glob-1"))
	router.Put("/admin/settings/*", testHandler("admin-glob-2"))
	router.Put("/users/:id", testHandler("user-id"))
	router.Put("/users/:id/favorite*", testHandler("user-favorites"))
	router.Put("/users/AB*", testHandler("user-glob-1"))

	assertRouter(router, "PUT", "/", "root")

	assertRouter(router, "PUT", "/admin", "admin-glob-1")
	assertRouter(router, "PUT", "/admin/a", "admin-glob-1")
	assertRouter(router, "PUT", "/admin/about", "admin-glob-1")
	assertRouter(router, "PUT", "/admin/about/123", "admin-glob-1")

	assertRouter(router, "PUT", "/admin/settings", "admin-glob-2")
	assertRouter(router, "PUT", "/admin/settings/", "admin-glob-2")
	assertRouter(router, "PUT", "/admin/settings/aab", "admin-glob-2")

	assertRouter(router, "PUT", "/users/1", "user-id")
	assertRouter(router, "PUT", "/users/b/favorites", "user-favorites")
	assertRouter(router, "PUT", "/users/b/favorites/323", "user-favorites")

	assertRouter(router, "PUT", "/users/ab", "user-glob-1")
	assertRouter(router, "PUT", "/users/aBaa", "user-glob-1")
	assertRouter(router, "PUT", "/users/ab444/asds", "user-glob-1")

	assertRouterNotFound(router, "PUT", "/user")
	assertRouterNotFound(router, "PUT", "/admi")
	assertRouterNotFound(router, "PUT", "/users/233/other")
}

func Benchmark_Router(b *testing.B) {
	router := New(Configure())
	router.Get("/users", testHandler("get-users"))
	router.Post("/users/:id", testHandler("create-user"))
	router.Get("/users/:userId/likes/:id", testHandler("get-favoriate"))
	requests := []*http.Request{
		build.Request().Path("/404").Request,
		build.Request().Path("/users").Request,
		build.Request().Path("/users/499/likes/001a").Request,
		build.Request().Method("Post").Path("/users/943").Request,
	}
	b.ResetTimer()
	res := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(res, requests[i%len(requests)])
	}
}

func assertRouting(routePath, requestPath string, params ...string) {
	router := New(Configure())
	router.Get(routePath, func(out http.ResponseWriter, req *Request) {
		Expect(req.params.Len()).To.Equal(len(params) / 2)
		for i := 0; i < len(params); i += 2 {
			key := params[i]
			Expect(req.Param(key)).To.Equal(params[i+1])
		}
		out.WriteHeader(200)
		out.Write([]byte(routePath))
	})
	assertRouter(router, "GET", requestPath, routePath, params...)
}

func assertRouter(router *Router, method string, requestPath string, body string, params ...string) {
	res := httptest.NewRecorder()
	req := build.Request().Path(requestPath).Method(method).Request
	router.ServeHTTP(res, req)
	Expect(res.Code).To.Equal(200)
	Expect(string(res.Body.Bytes())).To.Equal(body)
}

func assertNotFound(routePath, method string, requestPath string) {
	router := New(Configure())
	router.Get(routePath, func(out http.ResponseWriter, req *Request) {
		out.WriteHeader(200)
	})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, build.Request().Method(method).Path(requestPath).Request)
	Expect(res.Code).To.Equal(404)
}

func assertRouterNotFound(router *Router, method string, requestPath string) {
	res := httptest.NewRecorder()
	router.ServeHTTP(res, build.Request().Method(method).Path(requestPath).Request)
	Expect(res.Code).To.Equal(404)
}

func testHandler(body string) func(out http.ResponseWriter, req *Request) {
	return func(out http.ResponseWriter, req *Request) {
		out.WriteHeader(200)
		out.Write([]byte(body))
	}
}
