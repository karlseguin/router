# Router

A fast standalone router for Go.

Using <https://github.com/cypriss/golang-mux-benchmark>, `Router` was at least 2.5x faster (and up to 382x faster).

## Usage

Create, instance, configure your routes, hook into Go's HTTP server:

```go
import (
  "github.com/karlseguin/router"
  "net/http"
  "log"
)

func main() {
  router := router.New(router.Configure())
  router.Get("/users", userList)
  router.Get("/users/:id", userShow)

  s := &http.Server{
    Handler:        router,
  }
  log.Fatal(s.ListenAndServe())
}

func userList(out http.ResponseWriter, req *router.Request) {
  ...
}

func userShow(out http.ResponseWriter, req *router.Request) {
  id := req.Params["id"]
  ...
}
```

Notice that `userList` and `userShow` take a `*router.Request` and **not** a `*http.Request`. This is to expose the `Params` field.

## 404

Specify a handler for not found requests:

```go
router.NotFound(notFound)


func notFound(out http.ResponseWriter, req *router.Request) {
  ...
}
```

A basic not found handler is used by default.

# Coming Soon
* Constraints on parameters
* Glob routes (e.g., /users*)
