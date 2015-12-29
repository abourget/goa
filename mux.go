package goa

import (
	"fmt"
	"net/http"
)

// VersionMux is the interface implemented by the version handler used by goa to lookup the
// appropriate ServeMux for incoming requests.
// The generated code calls SetVersionMux once per version defined in the API design.
type VersionMux interface {
	SetDefaultMux(ServeMux)
	SetMux(version string, mux ServeMux)
	GetRequestMux(*http.Request) ServeMux
}

// DefaultVersionMux determines the API version targetted by the incoming request by looking for a
// "X-API-Version" HTTP header and  - if not found - a "api_version" querystring value.
type DefaultVersionMux struct {
	defaultMux ServeMux
	muxes      map[string]ServeMux
}

// NoVersionMux is the VersionMux used when the API design defines no version information.
type NoVersionMux struct {
	mux ServeMux
}

// ServeMux is the interface implemented by each version mux.
type ServeMux interface {
	http.Handler
	Handle(method, path string, handle HandleFunc)
	Lookup(method, path string) (HandleFunc, Params, bool)
}

// HandleFunc provides the implementation for an API endpoint.
type HandleFunc func(http.ResponseWriter, *http.Request, Params)

// Params is a slice of key/value pairs representing the captured request path parameters.
type Params []Param

// Param is a single request path parameter.
type Param struct {
	// Key is the value of the wildcard used to capture the parameter, e.g. "bottleID".
	Key string
	// Value represents the parameter captured value.
	Value string
}

// SetDefaultMux sets the mux used when the incoming request has no version information.
func (m *DefaultVersionMux) SetDefaultMux(mux ServeMux) {
	m.defaultMux = mux
}

// SetVersionMux records the mux for future lookup with GetRequestMux.
func (m *DefaultVersionMux) SetVersionMux(version string, mux ServeMux) {
	m.muxes[version] = mux
}

// GetRequestMux returns the mux associated with the incoming request by looking first for a
// "X-API-Version" header and if not found a "api_version" querystring value.
func (m *DefaultVersionMux) GetRequestMux(req *http.Request) (ServeMux, error) {
	version := req.Header.Get("X-API-Version")
	if version == "" {
		version = req.URL.Query().Get("api_version")
	}
	if version == "" {
		if m.defaultMux == nil {
			return nil, fmt.Errorf("no version defined in request and no default mux")
		}
		return m.defaultMux, nil
	}
	mux, ok := m.muxes[version]
	if !ok {
		return nil, fmt.Errorf(`no mux registered with version "%s"`, version)
	}
	return mux, nil
}

// SetDefaultMux sets the only mux used by NoVersionMux.
func (m *NoVersionMux) SetDefaultMux(mux ServeMux) {
	m.mux = mux
}

// SetVersionMux panics.
func (m *NoVersionMux) SetVersionMux(version string, mux ServeMux) {
	panic("trying to register a version mux with a no version mux")
}

// GetRequestMux returns an error.
func (m *NoVersionMux) GetRequestMux(req *http.Request) (ServeMux, error) {
	return nil, fmt.Errorf("no version registered")
}
