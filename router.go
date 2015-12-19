package goa

import "net/http"

type Router interface {
	Handle(method, path string, handle Handle)
	Lookup(method, path string) (Handle, Params, bool)
	ServeFiles(path string, root http.FileSystem)
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}
type Handle func(http.ResponseWriter, *http.Request, Params)
type Params []Param
type Param struct {
	Key   string
	Value string
}
