package goa

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Context is the object that provides access to the underlying HTTP request and response state.
// Context implements http.ResponseWriter and also provides helper methods for writing HTTP responses.
// It also implements the context.Context interface described at http://blog.golang.org/context.
type Context struct {
	context.Context // Underlying context
	log.Logger      // Context logger
}

// key is the type used to store internal values in the context.
// Context provides typed accessor methods to these values.
type key int

const (
	reqKey key = iota
	respKey
	paramsKey
	payloadKey
	respWrittenKey
	respStatusKey
	respLenKey
)

// NewContext builds a goa context from the given context.Context and request state.
// If gctx is nil then context.Background is used instead.
func NewContext(gctx context.Context,
	req *http.Request,
	rw http.ResponseWriter,
	params url.Values,
	payload interface{}) *Context {

	if gctx == nil {
		gctx = context.Background()
	}
	gctx = context.WithValue(gctx, reqKey, req)
	gctx = context.WithValue(gctx, respKey, rw)
	gctx = context.WithValue(gctx, paramsKey, params)
	gctx = context.WithValue(gctx, payloadKey, payload)

	return &Context{Context: gctx}
}

// SetValue sets the value associated with key in the context.
// The value can be retrieved using the Value method.
// Note that this changes the underlying context.Context object and thus clients holding a reference
// to that won't be able to access the new value. It's probably a bad idea to hold a reference to
// the inner context anyway...
func (ctx *Context) SetValue(key, val interface{}) {
	ctx.Context = context.WithValue(ctx.Context, key, val)
}

// Request returns the underlying HTTP request.
func (ctx *Context) Request() *http.Request {
	r := ctx.Value(reqKey)
	if r != nil {
		return r.(*http.Request)
	}
	return nil
}

// ResponseWritten returns true if an HTTP response was written.
func (ctx *Context) ResponseWritten() bool {
	if wr := ctx.Value(respStatusKey); wr != nil {
		return true
	}
	return false
}

// ResponseStatus returns the response status if it was set via one of the context response
// methods (Respond, JSON, BadRequest, Bug), 0 otherwise.
func (ctx *Context) ResponseStatus() int {
	if is := ctx.Value(respStatusKey); is != nil {
		return is.(int)
	}
	return 0
}

// ResponseLength returns the response body length in bytes if the response was written to the
// context via one of the response methods (Respond, JSON, BadRequest, Bug), 0 otherwise.
func (ctx *Context) ResponseLength() int {
	if is := ctx.Value(respLenKey); is != nil {
		return is.(int)
	}
	return 0
}

// Get returns the param or querystring value with the given name.
func (ctx *Context) Get(name string) string {
	iparams := ctx.Value(paramsKey)
	params := iparams.(url.Values)
	return params.Get(name)
}

// GetMany returns the querystring values with the given name or nil if there aren't any.
func (ctx *Context) GetMany(name string) ([]string, bool) {
	iparams := ctx.Value(paramsKey)
	params := iparams.(url.Values)
	p, ok := params[name]
	return p, ok
}

// Payload returns the deserialized request body or nil if body is empty.
func (ctx *Context) Payload() interface{} {
	return ctx.Value(payloadKey)
}

// Respond writes the given HTTP status code and response body.
// This method should only be called once per request.
func (ctx *Context) Respond(code int, body []byte) error {
	ctx.WriteHeader(code)
	if _, err := ctx.Write(body); err != nil {
		return err
	}
	return nil
}

// JSON serializes the given body into JSON and sends a HTTP response with the given status code
// and JSON as body.
func (ctx *Context) JSON(code int, body interface{}) error {
	js, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return ctx.Respond(code, js)
}

// BadRequest sends a HTTP response with status code 400 and the given error as body.
func (ctx *Context) BadRequest(err *BadRequestError) error {
	return ctx.Respond(400, []byte(err.Error()))
}

// Bug sends a HTTP response with status code 500 and the given body.
// The body can be set using a format and substituted values a la fmt.Printf.
func (ctx *Context) Bug(format string, a ...interface{}) error {
	body := fmt.Sprintf(format, a...)
	return ctx.Respond(500, []byte(body))
}

// Header returns the response header. It implements the http.ResponseWriter interface.
func (ctx *Context) Header() http.Header {
	rw := ctx.Value(respKey)
	if rw != nil {
		return rw.(http.ResponseWriter).Header()
	}
	return nil
}

// WriteHeader writes the HTTP status code to the response. It implements the
// http.ResponseWriter interface.
func (ctx *Context) WriteHeader(code int) {
	rw := ctx.Value(respKey)
	if rw != nil {
		ctx.Context = context.WithValue(ctx.Context, respStatusKey, code)
		rw.(http.ResponseWriter).WriteHeader(code)
	}
}

// Write writes the HTTP response body. It implements the http.ResponseWriter interface.
func (ctx *Context) Write(body []byte) (int, error) {
	rw := ctx.Value(respKey)
	if rw != nil {
		ctx.Context = context.WithValue(ctx.Context, respLenKey, ctx.ResponseLength()+len(body))
		return rw.(http.ResponseWriter).Write(body)
	}
	return 0, fmt.Errorf("response writer not initialized")
}
