package integration

import (
	"context"
	"mime/multipart"
	"net/http"

	rest "github.com/emicklei/go-restful/v3"
	"github.com/ggicci/httpin"
	"github.com/ggicci/httpin/core"
)

// HttpinFilter converts to a FilterFunction.
func HttpinFilter(input any, opts ...core.Option) rest.FilterFunction {
	handler := httpin.NewInput(input, opts...)
	return func(req *rest.Request, resp *rest.Response, chain *rest.FilterChain) {
		req.Request = setURLVars(req.Request, req.PathParameters())
		next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			req.Request = r
			resp.ResponseWriter = rw
			chain.ProcessFilter(req, resp)
		})

		handler(next).ServeHTTP(resp.ResponseWriter, req.Request)
	}
}

type contextKey int

const (
	varsKey contextKey = iota
)

// Vars returns the route variables for the current request, if any.
func Vars(r *http.Request) map[string]string {
	if rv := r.Context().Value(varsKey); rv != nil {
		return rv.(map[string]string)
	}
	return nil
}

// setURLVars sets the URL variables for the given request, to be accessed via
// integration.Vars for testing route behaviour. Arguments are not modified, a shallow
// copy is returned.
func setURLVars(r *http.Request, val map[string]string) *http.Request {
	return requestWithVars(r, val)
}

func requestWithVars(r *http.Request, vars map[string]string) *http.Request {
	ctx := context.WithValue(r.Context(), varsKey, vars)
	return r.WithContext(ctx)
}

// MuxVarsFunc is integration.Vars
type MuxVarsFunc func(*http.Request) map[string]string

// UseHttpin registers a new directive executor which can extract values
// from `integration.Vars`, i.e. path variables.
//
// Usage:
//
//	import integration "github.com/vine-io/go-restful-openapi/integration"
//
//	func init() {
//	    integration.UseHttpin("path", integration.Vars)
//	}
func UseHttpin(name string, fnVars MuxVarsFunc) {
	core.RegisterDirective(
		name,
		core.NewDirectivePath((&httpinVarsExtractor{Vars: fnVars}).Execute),
		true,
	)
}

type httpinVarsExtractor struct {
	Vars MuxVarsFunc
}

func (h *httpinVarsExtractor) Execute(rtm *core.DirectiveRuntime) error {
	req := rtm.GetRequest()
	kvs := make(map[string][]string)

	for key, value := range h.Vars(req) {
		kvs[key] = []string{value}
	}

	extractor := &core.FormExtractor{
		Runtime: rtm,
		Form: multipart.Form{
			Value: kvs,
		},
	}
	return extractor.Extract()
}
