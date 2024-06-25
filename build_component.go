package restfulspec

import (
	"reflect"

	"github.com/emicklei/go-restful/v3"
	spec "github.com/getkin/kin-openapi/openapi3"
)

func buildSchemas(ws *restful.WebService, cfg Config) (schemas spec.Schemas) {
	schemas = spec.Schemas{}
	for _, each := range ws.Routes() {
		addSchemaFromRouteTo(each, cfg, &schemas)
	}
	return
}

func addSchemaFromRouteTo(r restful.Route, cfg Config, s *spec.Schemas) {
	builder := schemaBuilder{Schemas: s, Config: cfg}
	if r.ReadSample != nil {
		builder.addModel(reflect.TypeOf(r.ReadSample), "")
	}
	if r.WriteSample != nil {
		builder.addModel(reflect.TypeOf(r.WriteSample), "")
	}
	for _, v := range r.ResponseErrors {
		if v.Model == nil {
			continue
		}
		builder.addModel(reflect.TypeOf(v.Model), "")
	}
}
