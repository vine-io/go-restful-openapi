package restspec

import (
	"github.com/emicklei/go-restful/v3"
	spec "github.com/getkin/kin-openapi/openapi3"
)

// NewOpenAPIService returns a new WebService that provides the API documentation of all services
// conform the OpenAPI documentation specifcation.
func NewOpenAPIService(config Config) *restful.WebService {

	ws := new(restful.WebService)
	ws.Path(config.APIPath)
	ws.Produces(restful.MIME_JSON)
	if !config.DisableCORS {
		ws.Filter(enableCORS)
	}

	openapi := BuildOpenAPIV3(config)
	resource := specResource{openapi: openapi}
	ws.Route(ws.GET("/").To(resource.getOpenAPI))
	return ws
}

// BuildOpenAPIV3 returns a openapi object for all services' API endpoints.
func BuildOpenAPIV3(config Config) *OpenAPI {
	// collect paths and model definitions to build Swagger object.
	paths := &spec.Paths{}
	components := &spec.Components{
		Schemas: map[string]*spec.SchemaRef{},
	}

	for _, each := range config.WebServices {
		builds := buildPaths(each, config)
		for path, item := range builds.Map() {
			existingPathItem := paths.Find(path)
			if existingPathItem != nil {
				for _, r := range each.Routes() {
					_, patterns := sanitizePath(r.Path)
					*item = buildPathItem(each, r, *existingPathItem, patterns, config)
				}
			}
			paths.Set(path, item)
		}
		components.Schemas = buildSchemas(each, config)
	}
	openapi := &OpenAPI{
		OpenAPI:    "3.0.1",
		Components: components,
		Info:       &spec.Info{},
		Paths:      paths,
		Security:   []spec.SecurityRequirement{},
		Servers: spec.Servers{{
			URL: config.Host,
		}},
		Tags:         []*spec.Tag{},
		ExternalDocs: &spec.ExternalDocs{},
	}
	if config.PostBuildOpenAPIObjectHandler != nil {
		config.PostBuildOpenAPIObjectHandler(openapi)
	}
	return openapi
}

func enableCORS(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	if origin := req.HeaderParameter(restful.HEADER_Origin); origin != "" {
		// prevent duplicate header
		if len(resp.Header().Get(restful.HEADER_AccessControlAllowOrigin)) == 0 {
			resp.AddHeader(restful.HEADER_AccessControlAllowOrigin, origin)
		}
	}
	chain.ProcessFilter(req, resp)
}

// specResource is a REST resource to serve the Open-API spec.
type specResource struct {
	openapi *OpenAPI
}

func (s specResource) getOpenAPI(req *restful.Request, resp *restful.Response) {
	resp.WriteAsJson(s.openapi)
}
