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
  id := req.Param("id")
  ...
}
```

Notice that `userList` and `userShow` take a `*router.Request` and **not** a `*http.Request`. This is to expose the `Param` method.

## Methods

All methods used to setup a route expect two parameters:

* path string
* handler func(res http.ResponseWriter, req *router.Request)

The methods are:

* Get
* Post
* Put
* Delete
* Patch
* Purge
* Head
* Options

The convenience method `All` sets up a route for all methods. This can be overwritten by setting a specific handler **before** calling `All`:

```go
router.Get("/power", testHandler("get-9000"))
router.All("/power", testHandler("9000"))
```

## Prefix matches

A route that ends with a '*' will do a prefix match on the incoming URL:

```go
router.Delete("/users/*", users)
```

Parameter matches are evaluated before prefix matches are considered. Given:

```go
router.Put("/users/:id", userShow)
router.Put("/users/ad*", userDebug)
```

and a request to `/users/ad123`, `userShow` will be executed.  However, a request
to `/users/ad/333` will execute `userDebug`.

Prefix matching is case insensitive.

## Constraints
Constraints can be placed on parameters:

```go
route.Delete("/users/:id(^\\d+$)", userDelete)
```

You'll very likely always want to bind to the start and end (^ and $), but this isn't automated in order to give you the flexibility of doing a partial match.

The code will panic on an invalid regular expression.

## Postfixes
A parameter can be followed by a postfix value, such as an extension:

```go
route.Get("/users/:id:.json", showUser)
```

## 404

Specify a handler for not found requests:

```go
router.NotFound(notFound)


func notFound(out http.ResponseWriter, req *router.Request) {
  ...
}
```

A default not found handler is used if none is provided
