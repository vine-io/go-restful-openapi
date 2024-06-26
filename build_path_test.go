package restspec

import (
	"testing"

	"github.com/emicklei/go-restful/v3"
	spec "github.com/getkin/kin-openapi/openapi3"
)

func TestRouteToPath(t *testing.T) {
	description := "get the <strong>a</strong> <em>b</em> test\nthis is the test description"
	notes := "notes\nblah blah"

	ws := new(restful.WebService)
	ws.Path("/tests/{v}")
	ws.Param(ws.PathParameter("v", "value of v").DefaultValue("default-v"))
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_XML)
	ws.Route(ws.GET("/a/{b}").To(dummy).
		Doc(description).
		Metadata(KeyOpenAPITags, []string{"tests"}).
		AddExtension("x-restful-test", "test value").
		Notes(notes).
		Param(ws.PathParameter("i", "some integer param").DataType("integer").AllowableValues(map[string]string{"0": "0", "1": "1"}).DefaultValue("1")).
		Param(ws.PathParameter("on", "some boolean param").DataType("boolean").AllowableValues(map[string]string{"true": "true", "false": "false"}).DefaultValue("false")).
		Param(ws.PathParameter("b", "value of b").DefaultValue("default-b")).
		Param(ws.QueryParameter("q", "value of q").DefaultValue("default-q")).
		Returns(200, "list of a b tests", []Sample{}).
		DefaultReturns("default", Sample{}).
		Writes([]Sample{}))
	ws.Route(ws.GET("/a/{b}/{c:[a-z]+}/{d:[1-9]+}/e/{f:*}").To(dummy).
		Param(ws.PathParameter("b", "value of b").DefaultValue("default-b")).
		Param(ws.PathParameter("c", "with regex").DefaultValue("abc")).
		Param(ws.PathParameter("d", "with regex").DefaultValue("abcef")).
		Param(ws.PathParameter("f", "with regex")).
		Param(ws.QueryParameter("q", "value of q").DataType("string").DataFormat("date").
			DefaultValue("default-q").AllowMultiple(true)).
		Returns(200, "list of a b tests", []Sample{}).
		Writes([]Sample{}))

	p := buildPaths(ws, Config{})
	t.Log(asJSON(p))

	if op := p.Find("/tests/{v}/a/{b}").Get; (*op.Parameters[0].Value.Schema.Value.Type)[0] != "string" {
		t.Error("Parameter type is not set.")
	} else if op.Extensions["x-restful-test"] != "test value" {
		t.Error("Extensions not set.")
	} else if len(op.Tags) != 1 || op.Tags[0] != "tests" {
		t.Error("Metadata/openapi.tags not set.")
	}
	if exists := p.Find("/tests/{v}/a/{b}/{c}/{d}/e/{f}"); exists == nil {
		t.Error("Expected path to exist after it was sanitized.")
	}

	checkParamTypes(t, p)

	q, exists := getParameter(*p.Find("/tests/{v}/a/{b}/{c}/{d}/e/{f}"), "q")
	if !exists {
		t.Errorf("get parameter q failed")
	}
	if (*q.Value.Schema.Value.Type)[0] != "array" || (*q.Value.Schema.Value.Items.Value.Type)[0] != "string" || q.Value.Schema.Value.Format != "date" {
		t.Errorf("parameter q expected to be a date array")
	}

	if p.Find("/tests/{v}/a/{b}").Get.Description != notes {
		t.Errorf("GET description incorrect")
	}
	if p.Find("/tests/{v}/a/{b}").Get.Summary != "get the a b test\nthis is the test description" {
		t.Errorf("GET summary incorrect")
	}
	response := p.Find("/tests/{v}/a/{b}").Get.Responses.Status(200)
	if (*response.Value.Content.Get("application/xml").Schema.Value.Type)[0] != "array" {
		t.Errorf("response type incorrect")
	}
	defaultResponse := p.Find("/tests/{v}/a/{b}").Get.Responses.Default()
	if (*defaultResponse.Value.Description) != "default" {
		t.Errorf("default response description incorrect")
	}
	if response.Value.Content["application/xml"].Schema.Value.Items.Ref != "#/components/schemas/restspec.Sample" {
		t.Errorf("response element type incorrect")
	}

	// Test for patterns
	path := p.Find("/tests/{v}/a/{b}/{c}/{d}/e/{f}")
	checkPattern(t, *path, "c", "[a-z]+")
	checkPattern(t, *path, "d", "[1-9]+")
	checkPattern(t, *path, "v", "")
	checkPattern(t, *path, "f", ".*")
}

func checkParamTypes(t *testing.T, p spec.Paths) {
	q, exists := getParameter(*p.Find("/tests/{v}/a/{b}"), "i")
	if !exists {
		t.Errorf("get parameter 'i' failed")
	}
	for _, enum := range q.Value.Schema.Value.Enum {
		_, ok := enum.(int64)
		if !ok {
			t.Errorf("enum for param 'i' is not an int64 type, type received: %T", enum)
		}
	}
	_, ok := q.Value.Schema.Value.Default.(int64)
	if !ok {
		t.Errorf("default for param 'i' is not an int64 type, type received: %T", q.Value.Schema.Value.Default)
	}

	q, exists = getParameter(*p.Find("/tests/{v}/a/{b}"), "on")
	if !exists {
		t.Errorf("get parameter 'on' failed")
	}
	for _, enum := range q.Value.Schema.Value.Enum {
		_, ok = enum.(bool)
		if !ok {
			t.Errorf("enum for param 'on' is not a boolean type, type received: %T", enum)
		}
	}
	_, ok = q.Value.Schema.Value.Default.(bool)
	if !ok {
		t.Errorf("default for param 'on' is not a boolean type, type received: %T", q.Value.Schema.Value.Default)
	}
}

func getParameter(path spec.PathItem, name string) (*spec.ParameterRef, bool) {
	for _, param := range path.Get.Parameters {
		if param.Value.Name == name {
			return param, true
		}
	}
	return nil, false
}

func checkPattern(t *testing.T, path spec.PathItem, paramName string, pattern string) {
	param, exists := getParameter(path, paramName)
	if !exists {
		t.Errorf("Expected Parameter %s to exist", paramName)
	}
	if param.Value.Schema.Value.Pattern != pattern {
		t.Errorf("Expected pattern %s to equal %s", param.Value.Schema.Value.Pattern, pattern)
	}
}

func TestGoogleCustomMethods(t *testing.T) {
	ws := new(restful.WebService)
	ws.Path("/tests").Consumes(restful.MIME_JSON).Produces(restful.MIME_XML)
	ws.Route(ws.GET("/resource:validate").To(dummy).
		Doc("validate resource").
		Returns(200, "validate resource successfully", []Sample{}).
		Writes([]Sample{}))
	ws.Route(ws.POST("/resource/{resourceId}:init").To(dummy).
		Doc("init the specified resource").
		Returns(200, "init the specified resource successfully", []Sample{}).
		Returns(500, "internal server error", []Sample{}).
		Reads(Sample{}).
		Writes([]Sample{}))

	p := buildPaths(ws, Config{})
	t.Log(asJSON(p))

	path := p.Find("/tests/resource:validate")
	if path == nil {
		t.Error("Expected path to exist after it was sanitized.")
		return
	}
	if path.Get.Summary != "validate resource" {
		t.Errorf("GET description incorrect")
	}
	response := path.Get.Responses.Status(200)
	if (*response.Value.Content["application/xml"].Schema.Value.Type)[0] != "array" {
		t.Errorf("response type incorrect")
	}
	if response.Value.Content["application/xml"].Schema.Value.Items.Ref != "#/components/schemas/restspec.Sample" {
		t.Errorf("response element type incorrect")
	}
	path = p.Find("/tests/resource/{resourceId}:init")
	if path == nil {
		t.Error("Expected path to exist after it was sanitized.")
		return
	}
	if path.Post.Summary != "init the specified resource" {
		t.Errorf("POST description incorrect")
	}
	if exists := p.Find("/tests/resource/{resourceId}:init").Post.Responses.Status(500); exists == nil {
		t.Errorf("Response code 500 not added to spec.")
	}
	response = path.Post.Responses.Status(200)
	if (*response.Value.Content["application/xml"].Schema.Value.Type)[0] != "array" {
		t.Errorf("response type incorrect")
	}
	if response.Value.Content["application/xml"].Schema.Value.Items.Ref != "#/components/schemas/restspec.Sample" {
		t.Errorf("response element type incorrect")
	}
}

func TestRouteToPathForAllowableValues(t *testing.T) {
	description := "get the <strong>a</strong> <em>b</em> test\nthis is the test description"
	notes := "notes\nblah blah"

	allowedCheeses := map[string]string{"cheddar": "cheddar", "feta": "feta", "colby-jack": "colby-jack", "mozzerella": "mozzerella"}
	ws := new(restful.WebService)
	ws.Path("/tests/{v}")
	ws.Param(ws.PathParameter("v", "value of v").DefaultValue("default-v"))
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_XML)
	ws.Route(ws.GET("/a/{cheese}").To(dummy).
		Doc(description).
		Notes(notes).
		Param(ws.PathParameter("cheese", "value of cheese").DefaultValue("cheddar").AllowableValues(allowedCheeses)).
		Param(ws.QueryParameter("q", "value of q").DefaultValue("default-q")).
		Returns(200, "list of a b tests", []Sample{}).
		Writes([]Sample{}))

	p := buildPaths(ws, Config{})
	t.Log(asJSON(p))

	if (*p.Find("/tests/{v}/a/{cheese}").Get.Parameters[0].Value.Schema.Value.Type)[0] != "string" {
		t.Error("Parameter type is not set.")
	}

	path := p.Find("/tests/{v}/a/{cheese}")
	if path == nil {
		t.Error("Expected path to exist after it was sanitized.")
		return
	}

	if path.Get.Description != notes {
		t.Errorf("GET description incorrect")
	}
	if path.Get.Summary != "get the a b test\nthis is the test description" {
		t.Errorf("GET summary incorrect")
	}
	if len(path.Get.Parameters) != 3 {
		t.Errorf("Path expected to have 3 parameters but had %d", len(path.Get.Parameters))
	}
	if path.Get.Parameters[1].Value.Name != "cheese" || len(path.Get.Parameters[1].Value.Schema.Value.Enum) != len(allowedCheeses) {
		t.Errorf("Path parameter 'cheese' expected to have enum of allowable values of lenght %d but was of length %d", len(allowedCheeses), len(path.Get.Parameters[1].Value.Schema.Value.Enum))
	}
	for _, cheeseIface := range path.Get.Parameters[1].Value.Schema.Value.Enum {
		cheese := cheeseIface.(string)
		if _, found := allowedCheeses[cheese]; !found {
			t.Errorf("Path parameter 'cheese' had enum value of %s which was not found in the allowableValues map", cheese)
		}
	}

	response := path.Get.Responses.Status(200)
	if (*response.Value.Content["application/xml"].Schema.Value.Type)[0] != "array" {
		t.Errorf("response type incorrect")
	}
	if response.Value.Content["application/xml"].Schema.Value.Items.Ref != "#/components/schemas/restspec.Sample" {
		t.Errorf("response element type incorrect")
	}
}

func TestMultipleMethodsRouteToPath(t *testing.T) {
	ws := new(restful.WebService)
	ws.Path("/tests/a")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_XML)
	ws.Route(ws.GET("/a/b").To(dummy).
		Doc("get a b test").
		Returns(200, "list of a b tests", []Sample{}).
		Writes([]Sample{}))
	ws.Route(ws.POST("/a/b").To(dummy).
		Doc("post a b test").
		Returns(200, "list of a b tests", []Sample{}).
		Returns(500, "internal server error", []Sample{}).
		Reads(Sample{}).
		Writes([]Sample{}))

	p := buildPaths(ws, Config{})
	t.Log(asJSON(p))

	if p.Find("/tests/a/a/b").Get.Summary != "get a b test" {
		t.Errorf("GET summary incorrect")
	}
	if p.Find("/tests/a/a/b").Post.Summary != "post a b test" {
		t.Errorf("POST summary incorrect")
	}
	if exists := p.Find("/tests/a/a/b").Post.Responses.Status(500); exists == nil {
		t.Errorf("Response code 500 not added to spec.")
	}

	expectedRef := "#/components/schemas/restspec.Sample"
	postBodyparam := p.Find("/tests/a/a/b").Post.RequestBody
	postBodyRef := postBodyparam.Value.Content["application/json"].Schema.Ref
	if postBodyRef != expectedRef {
		t.Errorf("Expected: %s, Got: %s", expectedRef, postBodyRef)
	}

	if postBodyparam.Value.Content["application/json"].Schema.Value.Format != "" || postBodyparam.Value.Content["application/json"].Schema.Value.Type.Slice() != nil || postBodyparam.Value.Content["application/json"].Schema.Value.Default != nil {
		t.Errorf("Invalid parameter property is set on body property")
	}
}

func TestReadArrayObjectInBody(t *testing.T) {
	ws := new(restful.WebService)
	ws.Path("/tests/a")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_XML)

	ws.Route(ws.POST("/a/b").To(dummy).
		Doc("post a b test with array in body").
		Returns(200, "list of a b tests", []Sample{}).
		Returns(500, "internal server error", []Sample{}).
		Reads([]Sample{}).
		Writes([]Sample{}))

	p := buildPaths(ws, Config{})
	//t.Log(asJSON(p))

	postInfo := p.Find("/tests/a/a/b").Post

	if postInfo.Summary != "post a b test with array in body" {
		t.Errorf("POST description incorrect")
	}
	if exists := postInfo.Responses.Status(500); exists == nil {
		t.Errorf("Response code 500 not added to spec.")
	}
	// identify element model type in body array
	expectedItemRef := "#/components/schemas/restspec.Sample"
	postBody := postInfo.RequestBody.Value.Content["application/json"].Schema
	if postBody.Ref != "" {
		t.Errorf("you shouldn't have body Ref setting when using array in body!")
	}
	// check body array dy item ref
	postBodyitems := postBody.Value.Items.Ref
	if postBodyitems != expectedItemRef {
		t.Errorf("Expected: %s, Got: %s", expectedItemRef, expectedItemRef)
	}

	if postBody.Value.Format != "" || postBody.Value.Type.Slice() == nil || postBody.Value.Default != nil {
		t.Errorf("Invalid parameter property is set on body property")
	}
}

// TestWritesPrimitive ensures that if an operation returns a primitive, then it
// is used as such (and not a ref to a definition).
func TestWritesPrimitive(t *testing.T) {
	ws := new(restful.WebService)
	ws.Path("/tests/returns")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/primitive").To(dummy).
		Doc("post that returns a string").
		Returns(200, "primitive string", "(this is a string)").
		Writes("(this is a string)"))

	ws.Route(ws.POST("/custom").To(dummy).
		Doc("post that returns a custom structure").
		Returns(200, "sample object", Sample{}).
		Writes(Sample{}))

	p := buildPaths(ws, Config{})
	//t.Log(asJSON(p))

	// Make sure that the operation that returns a primitive type is correct.
	if pathInfo := p.Find("/tests/returns/primitive"); pathInfo == nil {
		t.Errorf("Could not find path")
	} else {
		postInfo := pathInfo.Post

		if postInfo.Summary != "post that returns a string" {
			t.Errorf("POST description incorrect")
		}
		response := postInfo.Responses.Status(200)
		if response.Value.Content["application/json"].Schema.Ref != "" {
			t.Errorf("Expected no ref; got: %s", response.Value.Content["application/json"].Schema.Ref)
		}
		if len(response.Value.Content["application/json"].Schema.Value.Type.Slice()) != 1 {
			t.Errorf("Expected exactly one type; got: %d", len(response.Value.Content["application/json"].Schema.Value.Type.Slice()))
		}
		if (*response.Value.Content["application/json"].Schema.Value.Type)[0] != "string" {
			t.Errorf("Expected a type of string; got: %s", (*response.Value.Content["application/json"].Schema.Value.Type)[0])
		}
	}

	// Make sure that the operation that returns a custom type is correct.
	if pathInfo := p.Find("/tests/returns/custom"); pathInfo == nil {
		t.Errorf("Could not find path")
	} else {
		postInfo := pathInfo.Post

		if postInfo.Summary != "post that returns a custom structure" {
			t.Errorf("POST description incorrect")
		}
		response := postInfo.Responses.Status(200)
		if response.Value.Content["application/json"].Schema.Ref != "#/components/schemas/restspec.Sample" {
			t.Errorf("Expected ref '#/components/schemas/restspec.Sample'; got: %s", response.Value.Content["application/json"].Schema.Ref)
		}
		if len(response.Value.Content["application/json"].Schema.Value.Type.Slice()) != 0 {
			t.Errorf("Expected exactly zero types; got: %d", len(response.Value.Content["application/json"].Schema.Value.Type.Slice()))
		}
	}
}

// TestWritesRawSchema ensures that if an operation returns a raw schema value, then it
// is used as such (and not a ref to a definition).
func TestWritesRawSchema(t *testing.T) {
	ws := new(restful.WebService)
	ws.Path("/tests/returns")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/raw").To(dummy).
		Doc("get that returns a file").
		Returns(200, "raw schema type", SchemaType{RawType: "file"}).
		Writes(SchemaType{RawType: "file"}))

	p := buildPaths(ws, Config{})
	t.Log(asJSON(p))

	// Make sure that the operation that returns a raw schema type is correct.
	if pathInfo := p.Find("/tests/returns/raw"); pathInfo == nil {
		t.Errorf("Could not find path")
	} else {
		getInfo := pathInfo.Get

		if getInfo == nil {
			t.Errorf("operation was not present")
		}
		if getInfo.Summary != "get that returns a file" {
			t.Errorf("GET description incorrect")
		}
		response := getInfo.Responses.Status(200)
		if response.Value.Content["application/json"].Schema.Ref != "" {
			t.Errorf("Expected no ref; got: %s", response.Value.Content["application/json"].Schema.Ref)
		}
		if len(response.Value.Content["application/json"].Schema.Value.Type.Slice()) != 1 {
			t.Errorf("Expected exactly one type; got: %d", len(response.Value.Content["application/json"].Schema.Value.Type.Slice()))
		}
		if (*response.Value.Content["application/json"].Schema.Value.Type)[0] != "file" {
			t.Errorf("Expected a type of file; got: %s", (*response.Value.Content["application/json"].Schema.Value.Type)[0])
		}
	}
}

// TestWritesRawSchemaWithFormat ensures that if an operation returns a raw schema value with the specified format, then it
// is used as such (and not a ref to a definition).
func TestWritesRawSchemaWithFormat(t *testing.T) {
	ws := new(restful.WebService)
	ws.Path("/tests/returns")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/raw_formatted").To(dummy).
		Doc("get that returns a file").
		Returns(200, "raw schema type", SchemaType{RawType: "string", Format: "binary"}).
		Writes(SchemaType{RawType: "string", Format: "binary"}))

	p := buildPaths(ws, Config{})
	t.Log(asJSON(p))

	// Make sure that the operation that returns a raw schema type is correct.
	if pathInfo := p.Find("/tests/returns/raw_formatted"); pathInfo == nil {
		t.Errorf("Could not find path")
	} else {
		getInfo := pathInfo.Get

		if getInfo == nil {
			t.Errorf("operation was not present")
		}
		if getInfo.Summary != "get that returns a file" {
			t.Errorf("GET description incorrect")
		}
		response := getInfo.Responses.Status(200)
		if response.Value.Content["application/json"].Schema.Ref != "" {
			t.Errorf("Expected no ref; got: %s", response.Value.Content["application/json"].Schema.Ref)
		}
		if len(response.Value.Content["application/json"].Schema.Value.Type.Slice()) != 1 {
			t.Errorf("Expected exactly one type; got: %d", len(response.Value.Content["application/json"].Schema.Value.Type.Slice()))
		}
		if !response.Value.Content["application/json"].Schema.Value.Type.Is("string") {
			t.Errorf("Expected a type of string; got: %s", response.Value.Content["application/json"].Schema.Value.Type.Slice()[0])
		}
		if response.Value.Content["application/json"].Schema.Value.Format != "binary" {
			t.Errorf("Expected a format of binary; got: %s", response.Value.Content["application/json"].Schema.Value.Format)
		}
	}
}

// TestReadAndWriteArrayBytesInBody ensures that if an operation reads []byte in body or returns []byte,
// then it is represented as "string" with "binary" format.
func TestReadAndWriteArrayBytesInBody(t *testing.T) {
	ws := new(restful.WebService)
	ws.Path("/tests/a")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_XML)

	binaryType := SchemaType{RawType: "string", Format: "binary"}
	ws.Route(ws.POST("/a/b").To(dummy).
		Doc("post a b test with array of bytes in body").
		Returns(200, "list of a b tests", binaryType).
		Returns(500, "internal server error", binaryType).
		Reads(binaryType).
		Writes(binaryType))

	p := buildPaths(ws, Config{})
	t.Log(asJSON(p))

	postInfo := p.Find("/tests/a/a/b").Post

	if postInfo.Summary != "post a b test with array of bytes in body" {
		t.Errorf("POST description incorrect")
	}
	postBody := postInfo.RequestBody.Value.Content["application/json"].Schema

	if got, want := postBody.Value.Type.Slice()[0], "string"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := postBody.Value.Format, "binary"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if postBody.Ref != "" {
		t.Errorf("you shouldn't have body Ref setting when using array in body!")
	}
	if postBody.Value.Format == "" || postBody.Value.Default != nil {
		t.Errorf("Invalid parameter property is set on body property")
	}

	if exists := postInfo.Responses.Status(200); exists == nil {
		t.Errorf("Response code 200 not added to spec.")
	}
	sch := postInfo.Responses.Status(200).Value.Content["application/xml"].Schema.Value
	if sch == nil {
		t.Errorf("Schema for Response code 200 not added to spec.")
		return
	}
	if got, want := sch.Type.Slice()[0], "string"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := sch.Format, "binary"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if exists := postInfo.Responses.Status(500); exists == nil {
		t.Errorf("Response code 500 not added to spec.")
	}
	sch1 := postInfo.Responses.Status(500).Value.Content["application/xml"].Schema.Value
	if sch1 == nil {
		t.Errorf("Schema for Response code 500 not added to spec.")
		return
	}
	if got, want := (*sch1.Type)[0], "string"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := sch1.Format, "binary"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
}
