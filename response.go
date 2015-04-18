package router

import (
	"io"
	"log"
	"net/http"
)

var (
	BadRequest  = Respond(400, nil)
	NotFound    = Respond(404, nil)
	ServerError = Respond(500, nil)
)

type KeyValue struct {
	key   string
	value string
}

type Response interface {
	Header(key, value string) Response
	WriteTo(http.ResponseWriter)
}

// A response with an inmemory body
type NormalResponse struct {
	status  int
	headers []KeyValue
	body    []byte
}

func (r *NormalResponse) Header(key, value string) Response {
	r.headers = append(r.headers, KeyValue{key, value})
	return r
}

func (r *NormalResponse) WriteTo(out http.ResponseWriter) {
	writeHeaders(out.Header(), r.headers)
	out.WriteHeader(r.status)
	out.Write(r.body)
}

type StreamResponse struct {
	status  int
	headers []KeyValue
	body    io.Reader
}

func (r *StreamResponse) Header(key, value string) Response {
	r.headers = append(r.headers, KeyValue{key, value})
	return r
}

func (r *StreamResponse) WriteTo(out http.ResponseWriter) {
	if closer, ok := r.body.(io.Closer); ok {
		defer closer.Close()
	}
	writeHeaders(out.Header(), r.headers)
	out.WriteHeader(r.status)
	io.Copy(out, r.body)
}

func writeHeaders(out http.Header, headers []KeyValue) {
	for i, l := 0, len(headers); i < l; i++ {
		kv := headers[i]
		out.Set(kv.key, kv.value)
	}
}

func Empty(status int) Response {
	return Respond(status, nil)
}

func Json(status int, body []byte) Response {
	r := &NormalResponse{status: status, body: body}
	return r.Header("Content-Type", "application/json")
}

func Respond(status int, body []byte) Response {
	return &NormalResponse{status: status, body: body}
}

func Stream(status int, body io.Reader) Response {
	return &StreamResponse{status: status, body: body}
}

func Wrap(action func(req *Request) Response) func(http.ResponseWriter, *Request) {
	return func(out http.ResponseWriter, req *Request) {
		response := action(req)
		if response == nil {
			log.Println("nil response for", req.URL.String())
			response = ServerError
		}
		response.WriteTo(out)
	}
}
