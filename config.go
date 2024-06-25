package restfulspec

import (
	"reflect"

	"github.com/emicklei/go-restful/v3"
	spec "github.com/getkin/kin-openapi/openapi3"
)

// MapSchemaFormatFunc can be used to modify typeName at definition time.
// To use it set the SchemaFormatHandler in the config.
type MapSchemaFormatFunc func(typeName string) string

// MapModelTypeNameFunc can be used to return the desired typeName for a given
// type. It will return false if the default name should be used.
// To use it set the ModelTypeNameHandler in the config.
type MapModelTypeNameFunc func(t reflect.Type) (string, bool)

// PostBuildOpenAPIObjectFunc can be used to change the creates OpenAPI Object
// before serving it. To use it set the PostBuildOpenAPIObjectHandler in the config.
type PostBuildOpenAPIObjectFunc func(s *spec.T)

// ComponentNameHandlerFunc generate name by this handler for definition without json tag.
// example: (for more, see file definition_name_test.go)
//
//		  field	      			 definition_name
//		  Name `json:"name"`  ->  name
//	   Name                ->  Name
//
// there are some example provided for use
//
//	DefaultNameHandler         GoRestfulComponent -> GoRestfulComponent (not changed)
//	LowerSnakeCasedNameHandler  GoRestfulComponent -> go_restful_component
//	LowerCamelCasedNameHandler  GoRestfulComponent -> goRestfulComponent
//	GoLowerCamelCasedNameHandler HTTPRestfulComponent -> httpRestfulComponent
type ComponentNameHandlerFunc func(string) string

// Config holds service api metadata.
type Config struct {
	// [optional] If set then set this field with the generated Swagger Object
	Host string
	// [optional] If set then set this field with the generated Swagger Object
	Schemes []string
	// WebServicesURL is a DEPRECATED field; it never had any effect in this package.
	WebServicesURL string
	// APIPath is the path where the JSON api is available, e.g. /apidocs.json
	APIPath string
	// api listing is constructed from this list of restful WebServices.
	WebServices []*restful.WebService
	// [optional] on default CORS (Cross-Origin-Resource-Sharing) is enabled.
	DisableCORS bool
	// Top-level API version. Is reflected in the resource listing.
	APIVersion string
	// [optional] If set, model builder should call this handler to get addition typename-to-swagger-format-field conversion.
	SchemaFormatHandler MapSchemaFormatFunc
	// [optional] If set, model builder should call this handler to retrieve the name for a given type.
	ModelTypeNameHandler MapModelTypeNameFunc
	// [optional] If set then call this function with the generated OpenAPI Object
	PostBuildOpenAPIObjectHandler PostBuildOpenAPIObjectFunc
	// [optional] If set then call handler's function for to generate name by this handler for definition without json tag,
	//   you can use you ComponentNameHandler, also, there are four ComponentNameHandler provided, see definition_name.go
	ComponentNameHandler ComponentNameHandlerFunc
}
