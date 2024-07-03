package restspec

import (
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful/v3"
	spec "github.com/getkin/kin-openapi/openapi3"
)

const (
	// KeyOpenAPITags is a Metadata key for a restful Route
	KeyOpenAPITags = "openapi.tags"

	KeySecurityJWT = "security.jwt"

	// ExtensionPrefix is the only prefix accepted for VendorExtensible extension keys
	ExtensionPrefix = "x-"

	arrayType     = "array"
	componentRoot = "#/components/schemas/"
)

// SchemaType is used to wrap any raw types
// For example, to return a "schema": "file" one can use
// Returns(http.StatusOK, http.StatusText(http.StatusOK), SchemaType{RawType: "file"})
type SchemaType struct {
	RawType string
	Format  string
}

func buildPaths(ws *restful.WebService, cfg Config) spec.Paths {
	p := spec.Paths{}
	for _, each := range ws.Routes() {
		path, patterns := sanitizePath(each.Path)
		existingPathItem := p.Find(path)
		if existingPathItem == nil {
			existingPathItem = &spec.PathItem{}
		}
		pathItem := buildPathItem(ws, each, *existingPathItem, patterns, cfg)
		p.Set(path, &pathItem)
	}
	return p
}

// sanitizePath removes regex expressions from named path params,
// since openapi only supports setting the pattern as a property named "pattern".
// Expressions like "/api/v1/{name:[a-z]}/" are converted to "/api/v1/{name}/".
// The second return value is a map which contains the mapping from the path parameter
// name to the extracted pattern
func sanitizePath(restfulPath string) (string, map[string]string) {
	openapiPath := ""
	patterns := map[string]string{}
	for _, fragment := range strings.Split(restfulPath, "/") {
		if fragment == "" {
			continue
		}
		if strings.HasPrefix(fragment, "{") && strings.Contains(fragment, ":") {
			split := strings.Split(fragment, ":")
			// skip google custom method like `resource/{resource-id}:customVerb`
			if !strings.Contains(split[0], "}") {
				fragment = split[0][1:]
				pattern := split[1][:len(split[1])-1]
				if pattern == "*" { // special case
					pattern = ".*"
				}
				patterns[fragment] = pattern
				fragment = "{" + fragment + "}"
			}
		}
		openapiPath += "/" + fragment
	}
	return openapiPath, patterns
}

func buildPathItem(ws *restful.WebService, r restful.Route, existingPathItem spec.PathItem, patterns map[string]string, cfg Config) spec.PathItem {
	op := buildOperation(ws, r, patterns, cfg)
	switch r.Method {
	case http.MethodGet:
		existingPathItem.Get = op
	case http.MethodPost:
		existingPathItem.Post = op
	case http.MethodPut:
		existingPathItem.Put = op
	case http.MethodDelete:
		existingPathItem.Delete = op
	case http.MethodPatch:
		existingPathItem.Patch = op
	case http.MethodOptions:
		existingPathItem.Options = op
	case http.MethodHead:
		existingPathItem.Head = op
	}
	return existingPathItem
}

func buildOperation(ws *restful.WebService, r restful.Route, patterns map[string]string, cfg Config) *spec.Operation {
	o := spec.NewOperation()
	o.OperationID = r.Operation
	o.Description = r.Notes
	o.Summary = stripTags(r.Doc)
	o.Deprecated = r.Deprecated
	if r.Metadata != nil {
		if tags, ok := r.Metadata[KeyOpenAPITags]; ok {
			if tagList, ok := tags.([]string); ok {
				o.Tags = tagList
			}
		}
		if tags, ok := r.Metadata[KeySecurityJWT]; ok {
			if security, ok := tags.(string); ok {
				o.Security = &spec.SecurityRequirements{{
					security: []string{},
				}}
			}
		}
	}

	if o.Extensions == nil {
		o.Extensions = map[string]interface{}{}
	}
	extractExtensions(&o.Extensions, r.ExtensionProperties)

	// collect any path parameters
	for _, param := range ws.PathParameters() {
		p := buildParameter(r, param, patterns[param.Data().Name], cfg)
		o.AddParameter(&p)
	}
	// route specific params
	for _, param := range r.ParameterDocs {
		p := buildParameter(r, param, patterns[param.Data().Name], cfg)
		switch p.In {
		case "body":
			content := map[string]*spec.MediaType{}
			for _, consume := range r.Consumes {
				content[consume] = &spec.MediaType{
					Schema: p.Schema,
				}
			}

			o.RequestBody = &spec.RequestBodyRef{
				Value: &spec.RequestBody{
					Content: content,
				},
			}
		default:
			o.AddParameter(&p)
		}
	}
	o.Responses = new(spec.Responses)
	for k, v := range r.ResponseErrors {
		rsp := buildResponse(v, cfg, r.Produces)
		o.AddResponse(k, &rsp)
	}
	if r.DefaultResponse != nil {
		rsp := buildResponse(*r.DefaultResponse, cfg, r.Produces)
		o.AddResponse(-1, &rsp)
	}
	if o.Responses.Len() == 0 {
		o.AddResponse(200, (&spec.Response{}).WithDescription(http.StatusText(http.StatusOK)))
	}
	return o
}

// stringAutoType picks the correct type when dataType is set. Otherwise, it automatically picks the correct type from
// an ambiguously typed string. Ex. numbers become int, true/false become bool, etc.
func stringAutoType(dataType, ambiguous string) interface{} {
	if ambiguous == "" {
		return nil
	}

	if dataType == "" || dataType == "integer" {
		if parsedInt, err := strconv.ParseInt(ambiguous, 10, 64); err == nil {
			return parsedInt
		}
	}

	if dataType == "" || dataType == "boolean" {
		if parsedBool, err := strconv.ParseBool(ambiguous); err == nil {
			return parsedBool
		}
	}

	return ambiguous
}

func extractExtensions(extensible *map[string]interface{}, extensions restful.ExtensionProperties) {
	if len(extensions.Extensions) > 0 {
		for key := range extensions.Extensions {
			if strings.HasPrefix(key, ExtensionPrefix) {
				if *extensible == nil {
					m := map[string]interface{}{}
					extensible = &m
				}
				(*extensible)[key] = extensions.Extensions[key]
			}
		}
	}
}

func buildParameter(r restful.Route, restfulParam *restful.Parameter, pattern string, cfg Config) spec.Parameter {
	p := spec.Parameter{
		Schema: &spec.SchemaRef{
			Value: &spec.Schema{},
		},
	}
	param := restfulParam.Data()
	p.In = asParamType(param.Kind)

	if param.AllowMultiple {
		// If the param is an array apply the validations to the items in it
		p.Schema.Value.Type = &spec.Types{arrayType}
		p.Schema.Value.Items = &spec.SchemaRef{
			Value: &spec.Schema{
				Type:    &spec.Types{param.DataType},
				Pattern: param.Pattern,
			},
		}
		if param.MaxLength != nil {
			p.Schema.Value.Items.Value.MinLength = uint64(*param.MinLength)
		}
		p.Schema.Value.Format = param.CollectionFormat
		if param.MinItems != nil {
			p.Schema.Value.MinItems = uint64(*param.MinItems)
		}
		if param.MaxItems != nil {
			p.Schema.Value.MaxItems = spec.Uint64Ptr(uint64(*param.MaxItems))
		}
		p.Schema.Value.UniqueItems = param.UniqueItems
	} else {
		// Otherwise, for non-arrays, apply the validations directly to the param
		p.Schema.Value.Type = &spec.Types{param.DataType}
		if param.MinLength != nil {
			p.Schema.Value.MinLength = uint64(*param.MinLength)
		}
		if param.MaxLength != nil {
			p.Schema.Value.MaxLength = spec.Uint64Ptr(uint64(*param.MaxLength))
		}
		p.Schema.Value.Min = param.Minimum
		p.Schema.Value.Max = param.Maximum
	}

	// Prefer PossibleValues over deprecated AllowableValues
	if numPossible := len(param.PossibleValues); numPossible > 0 {
		// init Enum to our known size and populate it
		p.Schema.Value.Enum = make([]interface{}, 0, numPossible)
		for _, value := range param.PossibleValues {
			p.Schema.Value.Enum = append(p.Schema.Value.Enum, stringAutoType((*p.Schema.Value.Type)[0], value))
		}
	} else {
		if numAllowable := len(param.AllowableValues); numAllowable > 0 {
			// If allowable values are defined, set the enum array to the sorted values
			allowableSortedKeys := make([]string, 0, numAllowable)
			for k := range param.AllowableValues {
				allowableSortedKeys = append(allowableSortedKeys, k)
			}

			// sort away
			sort.Strings(allowableSortedKeys)

			// init Enum to our known size and populate it
			p.Schema.Value.Enum = make([]interface{}, 0, numAllowable)
			for _, key := range allowableSortedKeys {
				p.Schema.Value.Enum = append(p.Schema.Value.Enum, stringAutoType((*p.Schema.Value.Type)[0], key))
			}
		}
	}

	p.Description = param.Description
	p.Name = param.Name
	p.Required = param.Required
	p.AllowEmptyValue = param.AllowEmptyValue

	if param.Kind == restful.PathParameterKind {
		p.Schema.Value.Pattern = pattern
	} else if !param.AllowMultiple {
		p.Schema.Value.Pattern = param.Pattern
	}
	st := reflect.TypeOf(r.ReadSample)
	if param.Kind == restful.BodyParameterKind && r.ReadSample != nil && param.DataType == st.String() {
		p.Schema = &spec.SchemaRef{Value: spec.NewSchema()}
		if st.Kind() == reflect.Array || st.Kind() == reflect.Slice {
			dataTypeName := keyFrom(st.Elem(), cfg)
			p.Schema.Value.Type = &spec.Types{arrayType}
			p.Schema.Value.Items = &spec.SchemaRef{
				Value: spec.NewArraySchema(),
			}
			isPrimitive := isPrimitiveType(dataTypeName)
			if isPrimitive {
				mapped := jsonSchemaType(dataTypeName)
				p.Schema.Value.Items.Value.Type = &spec.Types{mapped}
			} else {
				p.Schema.Value.Items.Ref = componentRoot + dataTypeName
			}
		} else if schemaType, ok := r.ReadSample.(SchemaType); ok {
			p.Schema.Value.Type = &spec.Types{schemaType.RawType}
			p.Schema.Value.Format = schemaType.Format
		} else {
			dataTypeName := keyFrom(st, cfg)
			p.Schema.Ref = componentRoot + dataTypeName
			p.Schema.Value = spec.NewSchema()
		}

	} else {
		if param.AllowMultiple {
			p.Schema.Value.Type = &spec.Types{arrayType}
			p.Schema.Value.Items = &spec.SchemaRef{Value: &spec.Schema{}}
			p.Schema.Value.Items.Value.Type = &spec.Types{param.DataType}
			p.Schema.Value.Items.Value.Format = param.CollectionFormat
		} else {
			p.Schema.Value.Type = &spec.Types{param.DataType}
		}
		p.Schema.Value.Default = stringAutoType(param.DataType, param.DefaultValue)
		p.Schema.Value.Format = param.DataFormat
	}

	if p.Extensions == nil {
		p.Extensions = map[string]interface{}{}
	}
	extractExtensions(&p.Extensions, param.ExtensionProperties)

	return p
}

func buildResponse(e restful.ResponseError, cfg Config, products []string) (r spec.Response) {
	r.Description = new(string)
	*r.Description = e.Message
	if e.Model != nil {
		st := reflect.TypeOf(e.Model)
		if st.Kind() == reflect.Ptr {
			// For pointer type, use element type as the key; otherwise we'll
			// endue with '#/components/schemas/*Type' which violates openapi spec.
			st = st.Elem()
		}
		schema := &spec.SchemaRef{Value: &spec.Schema{}}
		if st.Kind() == reflect.Array || st.Kind() == reflect.Slice {
			modelName := keyFrom(st.Elem(), cfg)
			schema.Value.Type = &spec.Types{arrayType}
			schema.Value.Items = &spec.SchemaRef{
				Value: spec.NewArraySchema(),
			}
			isPrimitive := isPrimitiveType(modelName)
			if isPrimitive {
				mapped := jsonSchemaType(modelName)
				schema.Value.Items.Value.Type = &spec.Types{mapped}
			} else {
				schema.Value.Items.Ref = componentRoot + modelName
			}
		} else {
			modelName := keyFrom(st, cfg)
			if schema.Value.Type == nil {
				schema.Value.Type = &spec.Types{}
			}
			if isPrimitiveType(modelName) {
				// If the response is a primitive type, then don't reference any definitions.
				// Instead, set the schema's "type" to the model name.
				*schema.Value.Type = append(*(schema.Value.Type), modelName)
			} else if schemaType, ok := e.Model.(SchemaType); ok {
				*schema.Value.Type = append(*(schema.Value.Type), schemaType.RawType)
				schema.Value.Format = schemaType.Format
			} else {
				modelName = keyFrom(st, cfg)
				schema.Ref = componentRoot + modelName
			}
		}

		contents := map[string]*spec.MediaType{}
		for _, product := range products {
			contents[product] = &spec.MediaType{
				Schema: schema,
			}
		}
		r.WithContent(contents)
	}

	if len(e.Headers) > 0 {
		r.Headers = make(map[string]*spec.HeaderRef, len(e.Headers))
		for k, v := range e.Headers {
			headerRef := buildHeader(v)
			r.Headers[k] = &headerRef
		}
	}

	if r.Extensions == nil {
		r.Extensions = map[string]interface{}{}
	}
	extractExtensions(&r.Extensions, e.ExtensionProperties)
	return r
}

// buildHeader builds a specification header structure from restful.Header
func buildHeader(header restful.Header) spec.HeaderRef {
	responseHeader := spec.HeaderRef{
		Value: &spec.Header{
			Parameter: spec.Parameter{
				Schema: &spec.SchemaRef{
					Value: &spec.Schema{},
				},
			},
		},
	}
	responseHeader.Value.In = header.Type
	responseHeader.Value.Description = header.Description
	responseHeader.Value.Schema.Value.Format = header.Format
	responseHeader.Value.Schema.Value.Default = header.Default

	// If type is "array" items field is required
	if header.Type == arrayType {
		responseHeader.Value.Schema.Value.Format = header.CollectionFormat
		responseHeader.Value.Schema.Value.Items = buildHeadersItems(header.Items)
	}

	return responseHeader
}

// buildHeadersItems builds
func buildHeadersItems(items *restful.Items) *spec.SchemaRef {
	responseItems := &spec.SchemaRef{Value: &spec.Schema{}}
	responseItems.Value.Format = items.Format
	responseItems.Value.Type = &spec.Types{arrayType}
	responseItems.Value.Default = items.Default
	responseItems.Value.Format = items.CollectionFormat
	if items.Items != nil {
		responseItems.Value.Items = buildHeadersItems(items.Items)
	}

	return responseItems
}

// stripTags takes a snippet of HTML and returns only the text content.
// For example, `<b>&lt;Hi!&gt;</b> <br>` -> `&lt;Hi!&gt; `.
func stripTags(html string) string {
	re := regexp.MustCompile("<[^>]*>")
	return re.ReplaceAllString(html, "")
}

func isPrimitiveType(modelName string) bool {
	if len(modelName) == 0 {
		return false
	}
	return strings.Contains("uint uint8 uint16 uint32 uint64 int int8 int16 int32 int64 float32 float64 bool string byte rune time.Time time.Duration", modelName)
}

func jsonSchemaType(modelName string) string {
	schemaMap := map[string]string{
		"uint":   "integer",
		"uint8":  "integer",
		"uint16": "integer",
		"uint32": "integer",
		"uint64": "integer",

		"int":   "integer",
		"int8":  "integer",
		"int16": "integer",
		"int32": "integer",
		"int64": "integer",

		"byte":          "integer",
		"float64":       "number",
		"float32":       "number",
		"bool":          "boolean",
		"time.Time":     "string",
		"time.Duration": "integer",
	}
	mapped, ok := schemaMap[modelName]
	if !ok {
		return modelName // use as is (custom or struct)
	}
	return mapped
}
