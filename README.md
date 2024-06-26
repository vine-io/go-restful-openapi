# go-restful-openapi

[![Build Status](https://travis-ci.org/vine-io/go-restful-openapi.png)](https://travis-ci.org/vine-io/go-restful-openapi)
[![GoDoc](https://godoc.org/github.com/vine-io/go-restful-openapi?status.svg)](https://godoc.org/github.com/vine-io/go-restful-openapi)

[openapi](https://www.openapis.org) extension to the go-restful package, targeting [version 3.0](https://github.com/OAI/OpenAPI-Specification)

## The following Go field tags are translated to OpenAPI equivalents
- description
- minimum
- maximum
- optional ( if set to "true" then it is not listed in `required`)
- unique
- modelDescription
- type (overrides the Go type String())
- enum
- readOnly

See TestThatExtraTagsAreReadIntoModel for examples.

## dependencies

- [go-restful](https://github.com/emicklei/go-restful)
- [kin-openapi](github.com/getkin/kin-openapi/openapi3)


## Go modules

    import (
        restfulspec "github.com/vine-io/go-restful-openapi"
	    restful "github.com/emicklei/go-restful/v3"
    )


© 2024, MIT License. Contributions welcome.