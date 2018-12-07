# Swagger Generator (swgen)

[![Build Status](https://travis-ci.org/swaggest/swgen.svg?branch=master)](https://travis-ci.org/swaggest/swgen)
[![Coverage Status](https://codecov.io/gh/swaggest/swgen/branch/master/graph/badge.svg)](https://codecov.io/gh/swaggest/swgen)
[![GoDoc](https://godoc.org/github.com/swaggest/swgen?status.svg)](https://godoc.org/github.com/swaggest/swgen)

Swagger Generator is a library which helps to generate [Swagger Specification](http://swagger.io/specification/) in JSON format on-the-fly.

## Installation

You can use `go get` to install the `swgen` package

    go get github.com/swaggest/swgen

Then import it into your own code

```go
import "github.com/swaggest/swgen"
```

## Example

```go
package main

import (
    "fmt"

    "github.com/swaggest/swgen"
)

// PetsRequest defines all params for /pets request
type PetsRequest struct {
    Tags  []string `schema:"tags" in:"query" required:"-" description:"tags to filter by"`
    Limit int32    `schema:"limit" in:"query" required:"-" description:"maximum number of results to return"`
}

// Pet contains information of a pet
type Pet struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
    Tag  string `json:"tag"`
}

func main() {
	gen := swgen.NewGenerator()
	gen.SetHost("petstore.swagger.io").SetBasePath("/api")
	gen.SetInfo("Swagger Petstore (Simple)", "A sample API that uses a petstore as an example to demonstrate features in the swagger-2.0 specification", "http://helloreverb.com/terms/", "2.0")
	gen.SetLicense("MIT", "http://opensource.org/licenses/MIT")
	gen.SetContact("Swagger API team", "http://swagger.io", "foo@example.com")
	gen.AddSecurityDefinition("BasicAuth", swgen.SecurityDef{Type: swgen.SecurityBasicAuth})

	pathInf := swgen.PathItemInfo{
		Path:        "/pets",
		Method:      "GET",
		Title:       "findPets",
		Description: "Returns all pets from the system that the user has access to",
		Tag:         "v1",
		Deprecated:  false,
		Security:    []string{"BasicAuth"},
		Request:     new(PetsRequest), // request object
		Response:    new([]Pet),       // response object
	}
	pathInf.AddExtendedField("x-example", "example")

	gen.SetPathItem(pathInf)

	// extended field
	gen.AddExtendedField("x-uppercase-version", true)

	docData, _ := gen.GenDocument()
	fmt.Println(string(docData))

	// output:
	// {"swagger":"2.0","info":{"title":"Swagger Petstore (Simple)","description":"A sample API that uses a petstore as an example to demonstrate features in the swagger-2.0 specification","termsOfService":"http://helloreverb.com/terms/","contact":{"name":"Swagger API team","url":"http://swagger.io","email":"foo@example.com"},"license":{"name":"MIT","url":"http://opensource.org/licenses/MIT"},"version":"2.0"},"host":"petstore.swagger.io","basePath":"/api","schemes":["http","https"],"paths":{"/pets":{"get":{"tags":["v1"],"summary":"findPets","description":"Returns all pets from the system that the user has access to","parameters":[{"description":"tags to filter by","type":"array","name":"tags","in":"query","items":{"type":"string"},"collectionFormat":"multi"},{"description":"maximum number of results to return","type":"integer","format":"int32","name":"limit","in":"query"}],"responses":{"200":{"description":"OK","schema":{"type":"array","items":{"$ref":"#/definitions/Pet"}}}},"security":[{"BasicAuth":[]}],"x-example":"example"}}},"definitions":{"Pet":{"type":"object","properties":{"id":{"type":"integer","format":"int64"},"name":{"type":"string"},"tag":{"type":"string"}}}},"securityDefinitions":{"BasicAuth":{"type":"basic"}},"x-uppercase-version":true}
}
```

## License

Distributed under the Apache License, version 2.0.
Please see license file in code for more details.
