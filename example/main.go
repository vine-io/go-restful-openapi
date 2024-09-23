package main

import (
	"fmt"
	"log"
	"net/http"

	rest "github.com/emicklei/go-restful/v3"
	spec "github.com/getkin/kin-openapi/openapi3"
	"github.com/ggicci/httpin"

	restspec "github.com/vine-io/go-restful-openapi"
	"github.com/vine-io/go-restful-openapi/example/apis"
	"github.com/vine-io/go-restful-openapi/integration"
)

func init() {
	integration.UseHttpin("path", integration.Vars)
}

func bodyLogFilter(req *rest.Request, resp *rest.Response, chain *rest.FilterChain) {

	//inBody, err := io.ReadAll(req.Request.Body)
	//if err != nil {
	//	resp.WriteError(400, err)
	//	return
	//}
	//req.Request.Body = ioutil.NopCloser(bytes.NewReader(inBody))

	//c := NewResponseCapture(resp.ResponseWriter)
	//resp.ResponseWriter = c

	chain.ProcessFilter(req, resp)

	//log.Println("Request body:", string(inBody))
	//log.Println("Response body:", string(c.Bytes()))
}

// rest

type UserResource struct {
	// normally one would use DAO (data access object)
	users map[string]apis.User
}

func (u UserResource) WebService() *rest.WebService {
	ws := new(rest.WebService)
	ws.
		Path("/users").
		Consumes(rest.MIME_JSON, rest.MIME_XML).
		Produces(rest.MIME_JSON, rest.MIME_XML) // you can specify this per route as well

	ws.Filter(func(req *rest.Request, rsp *rest.Response, chain *rest.FilterChain) {
		chain.ProcessFilter(req, rsp)
	})
	tags := []string{"users"}

	ws.Route(ws.GET("/").To(u.findAllUsers).
		// docs
		Doc("get all users").
		Metadata(restspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter("gender", "identifier of the user").DataType("string")).
		Param(ws.HeaderParameter("token", "user token").DataType("string")).
		Returns(200, "OK", []apis.User{}))

	ws.Route(ws.GET("/{id}").To(u.findUser).
		// docs
		Operation("User.List").
		Doc("get a user").
		Param(ws.PathParameter("id", "identifier of the user").DataType("integer").DefaultValue("1")).
		Metadata(restspec.KeyOpenAPITags, tags).
		Metadata(restspec.KeySecurityJWT, "bearerAuth").
		Writes(apis.User{}). // on the response
		Returns(200, "OK", apis.User{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.PATCH("/{id}").
		Filter(integration.HttpinFilter(apis.UpdateUserInput{})).
		To(u.updateUser).
		// docs
		Doc("update a user").
		//Param(ws.PathParameter("id", "identifier of the user").DataType("string")).
		Metadata(restspec.KeyOpenAPITags, tags).
		//Reads(apis.UserPatch{}),
		Do(restspec.ReadSample(apis.UpdateUserInput{})),
	)
	//Reads(apis.UpdateUserInput{})) // from the request

	ws.Route(ws.POST("").To(u.createUser).
		// docs
		Doc("create a user").
		Metadata(restspec.KeyOpenAPITags, tags).
		Reads(apis.User{}).
		Returns(200, "OK", apis.User{})) // from the request

	ws.Route(ws.DELETE("/{id}").To(u.removeUser).
		// docs
		Doc("delete a user").
		Metadata(restspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter("id", "identifier of the user").DataType("string")))

	return ws
}

// GET http://localhost:8080/users
func (u UserResource) findAllUsers(request *rest.Request, response *rest.Response) {

	list := []apis.User{}
	for _, each := range u.users {
		list = append(list, each)
	}
	response.WriteEntity(list)
}

// GET http://localhost:8080/users/1
func (u UserResource) findUser(request *rest.Request, response *rest.Response) {
	id := request.PathParameter("id")
	usr := u.users[id]
	if len(usr.ID) == 0 {
		response.WriteErrorString(http.StatusNotFound, "User could not be found.")
	} else {
		response.WriteEntity(usr)
	}
}

// PUT http://localhost:8080/users/1
func (u *UserResource) updateUser(request *rest.Request, response *rest.Response) {
	input := request.Request.Context().Value(httpin.Input).(*apis.UpdateUserInput)
	fmt.Printf("%#v\n", input)

	usr := new(apis.User)
	err := request.ReadEntity(&usr)
	if err == nil {
		u.users[usr.ID] = *usr
		response.WriteEntity(usr)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

// PUT http://localhost:8080/users/1
func (u *UserResource) createUser(request *rest.Request, response *rest.Response) {
	usr := apis.User{ID: request.PathParameter("id")}
	err := request.ReadEntity(&usr)
	if err == nil {
		u.users[usr.ID] = usr
		response.WriteHeaderAndEntity(http.StatusCreated, usr)
	} else {
		response.WriteError(http.StatusInternalServerError, err)
	}
}

// DELETE http://localhost:8080/users/1
func (u *UserResource) removeUser(request *rest.Request, response *rest.Response) {
	id := request.PathParameter("id")
	delete(u.users, id)
}

func main() {
	root := rest.NewContainer()
	u := UserResource{map[string]apis.User{}}
	root.Add(u.WebService())

	config := restspec.Config{
		WebServices:                   root.RegisteredWebServices(), // you control what services are visible
		APIPath:                       "/openapi.json",
		PostBuildOpenAPIObjectHandler: enrichOpenAPIObject,
		//ModelTypeNameHandler: func(t reflect.Type) (string, bool) {
		//	// fmt.Println(t.String(), t.Align(), t.FieldAlign())
		//	pkg := strings.ReplaceAll(t.PkgPath(), "/", "_")
		//	return pkg + "." + t.Name(), true
		//},
		Host: "http://localhost:8081",
	}
	root.Add(restspec.NewOpenAPIService(config))

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs/?url=http://localhost:8080/apidocs.json
	root.Handle("/apidocs/", http.StripPrefix("/apidocs/", http.FileServer(http.Dir("../testdata/swagger"))))

	// Optionally, you may need to enable CORS for the UI to work.
	cors := rest.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		CookiesAllowed: false,
		Container:      root,
	}
	root.Filter(cors.Filter)

	log.Printf("Get the API using http://localhost:8081/apidocs.json")
	log.Printf("Open Swagger UI using http://localhost:8081/apidocs/?url=http://localhost:8080/apidocs.json")
	log.Fatal(http.ListenAndServe(":8081", root))
}

func enrichOpenAPIObject(swo *restspec.OpenAPI) {

	swo.Info = &spec.Info{
		Title:       "UserService",
		Description: "Resource for managing Users",
		Contact: &spec.Contact{
			Name:  "john",
			Email: "john@doe.rp",
			URL:   "http://johndoe.org",
		},
		License: &spec.License{
			Name: "MIT",
			URL:  "http://mit.org",
		},
		Version: "1.0.0",
	}
	swo.Tags = spec.Tags{
		{
			Name:        "users",
			Description: "Managing users",
		},
	}

	swo.Components.SecuritySchemes = spec.SecuritySchemes{
		"bearerAuth": &spec.SecuritySchemeRef{
			Value: &spec.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		},
	}
}
