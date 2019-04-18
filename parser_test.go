package swgen

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Person struct {
	Name       *PersonName `json:"name"`
	SecondName PersonName  `json:"second_name" required:"true"`
	Age        uint        `json:"age"`
	Children   []Person    `json:"children"`
	Tags       []string    `json:"tags"`
	Weight     float64     `json:"weight"`
	Active     bool        `json:"active"`
	Balance    float32     `json:"balance"`
}

type PersonName struct {
	First    string `json:"first_name"`
	Middle   string `json:"middle_name"`
	Last     string `json:"last_name"`
	Nickname string `query:"-"`
	_        string
}

type Employee struct {
	Person
	Salary float64 `json:"salary"`
}

type Project struct {
	ID      uint        `json:"id"`
	Name    string      `json:"name"`
	Manager interface{} `json:"manager"`
}

// PreferredWarehouseRequest is request object of get preferred warehouse handler
type PreferredWarehouseRequest struct {
	Items              []string `query:"items" description:"List of simple sku"`
	IDCustomerLocation uint64   `query:"id_customer_location" description:"-"`
}

func TestResetDefinitions(t *testing.T) {
	ts := &Person{}
	gen := NewGenerator()
	gen.ParseDefinition(ts)

	assert.NotEqual(t, 0, len(gen.definitions))
	gen.ResetDefinitions()
	assert.Equal(t, 0, len(gen.definitions))
}

func TestParseDefinition(t *testing.T) {
	ts := &Person{}
	NewGenerator().ParseDefinition(ts)
}

func TestParseDefinitionEmptyInterface(t *testing.T) {
	var ts interface{}
	gen := NewGenerator()
	gen.ParseDefinition(&ts)
}

func TestParseDefinitionNonEmptyInterface(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Panic expected for non-empty interface")
		}
	}()

	var ts interface {
		Test()
	}

	NewGenerator().ParseDefinition(&ts)
}

func TestParseDefinitionWithEmbeddedStruct(t *testing.T) {
	ts := &Employee{}
	tt := reflect.TypeOf(ts)

	gen := NewGenerator()
	gen.ParseDefinition(ts)

	if typeDef, found := gen.getDefinition(tt); !found {
		assert.True(t, found)
	} else {
		propertiesCount := len(typeDef.Properties)
		expectedPropertiesCount := 9
		assert.Equalf(t, expectedPropertiesCount, propertiesCount, "%#v", typeDef.Properties)
	}
}

func TestParseDefinitionWithEmbeddedInterface(t *testing.T) {
	p := &Project{Manager: new(Employee)}
	tt := reflect.TypeOf(p)

	gen := NewGenerator()
	gen.ParseDefinition(p)

	if typeDef, found := gen.getDefinition(tt); !found {
		assert.True(t, found)
	} else {
		assert.Equal(t, "#/definitions/Employee", typeDef.Properties["manager"].Ref)
	}
}

func TestParseDefinitionString(t *testing.T) {
	typeDef := NewGenerator().ParseDefinition("string")
	name := typeDef.TypeName
	assert.Equal(t, "string", name)
}

func TestParseDefinitionArray(t *testing.T) {
	type Names []string
	typeDef := NewGenerator().ParseDefinition(Names{})

	assert.Equal(t, "Names", typeDef.TypeName)

	// re-parse with pointer input
	// should get from definition list
	NewGenerator().ParseDefinition(&Names{})

	// try to parse a named map
	type MapList map[string]string
	NewGenerator().ParseDefinition(&MapList{})

	// named array of object
	type Person struct{}
	type Persons []*Person
	NewGenerator().ParseDefinition(&Persons{})
}

func TestParseParameter(t *testing.T) {
	p := &PreferredWarehouseRequest{}
	name, params := NewGenerator().ParseParameters(p)

	assert.Equal(t, "PreferredWarehouseRequest", name)
	assert.Len(t, params, 2)
}

func TestParseParameterError(t *testing.T) {
	assert.Panics(t, func() { NewGenerator().ParseParameters(true) })
}

//
// test and data for TestSetPathItem
//

func TestSetPathItem(t *testing.T) {
	h := &testHandler{}

	gen := NewGenerator()
	methods := []string{"GET", "POST", "HEAD", "PUT", "OPTIONS", "DELETE", "PATCH"}

	for _, method := range methods {
		info := PathItemInfo{
			Path:        "/v1/test/handler",
			Title:       "TestHandler",
			Description: fmt.Sprintf("This is just a test handler with %s request", method),
			Method:      method,
			Request:     h.GetRequestBuffer(method),
			Response:    h.GetResponseBuffer(method),
		}
		gen.SetPathItem(info)
	}

	assert.NotEqual(t, 0, len(gen.paths))
	gen.ResetPaths()
	assert.Equal(t, 0, len(gen.paths))
}

// testHandler can handle POST and GET request
type testHandler struct{}

func (th *testHandler) GetName() string {
	return "TestHandle"
}

func (th *testHandler) GetDescription() string {
	return "This handler for test ParsePathItem"
}

func (th *testHandler) GetVersion() string {
	return "v1"
}

func (th *testHandler) GetRoute() string {
	return "/test/handler"
}

func (th *testHandler) GetRequestBuffer(_ string) interface{} {
	return &PersonName{}
}

func (th *testHandler) GetResponseBuffer(method string) interface{} {
	if method == "GET" {
		return nil
	}

	return &PreferredWarehouseRequest{}
}

func (th *testHandler) GetBodyBuffer() interface{} {
	return &Person{}
}

func (th *testHandler) HandlePost(_ interface{}, _ interface{}) (response interface{}, err error) {
	// yes, I can handle a POST request
	return
}

func (th *testHandler) HandleGet(_ interface{}) (response interface{}, err error) {
	// yes, I can handle a GET request
	return
}

type custom string

func (custom) SwaggerDef() SwaggerData {
	d := SwaggerData{}
	d.Description = "A custom string"
	d.Type = "string"
	d.Pattern = "^[a-z]{4}$"
	return d
}

type bodyWithCustom struct {
	ID int    `json:"id"`
	C  custom `json:"c"`
}

type paramsWithCustom struct {
	ID int    `query:"id" required:"true"`
	C  custom `query:"c" required:"true"`
}

func TestSwaggerDef(t *testing.T) {
	gen := NewGenerator()

	gen.SetPathItem(PathItemInfo{
		Method:   http.MethodPost,
		Path:     "/bla",
		Request:  new(bodyWithCustom),
		Consumes: []string{"application/json"},
		Produces: []string{"text/csv"},
	})

	gen.SetPathItem(PathItemInfo{
		Method:  http.MethodGet,
		Path:    "/bla",
		Request: new(paramsWithCustom),
	})

	swg, err := gen.GenDocument()
	assert.NoError(t, err)
	expected := `
{
  "swagger": "2.0",
  "info": {
    "title": "",
    "description": "",
    "termsOfService": "",
    "contact": {
      "name": ""
    },
    "license": {
      "name": ""
    },
    "version": ""
  },
  "basePath": "/",
  "schemes": [
    "http",
    "https"
  ],
  "paths": {
    "/bla": {
      "get": {
        "summary": "",
        "description": "",
        "parameters": [
          {
            "type": "integer",
            "format": "int32",
            "name": "id",
            "in": "query",
            "required": true
          },
          {
            "description": "A custom string",
            "type": "string",
            "pattern": "^[a-z]{4}$",
            "name": "c",
            "in": "query",
            "required": true
          }
        ],
        "responses": {
          "204": {
            "description": "No Content"
          }
        }
      },
      "post": {
        "summary": "",
        "produces": ["text/csv"],
        "consumes": ["application/json"],
        "description": "",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "schema": {
              "$ref": "#/definitions/bodyWithCustom"
            },
            "required": true
          }
        ],
        "responses": {
          "204": {
            "description": "No Content"
          }
        }
      }
    }
  },
  "definitions": {
    "bodyWithCustom": {
      "type": "object",
      "properties": {
        "c": {
          "$ref": "#/definitions/custom"
        },
        "id": {
          "type": "integer",
          "format": "int32"
        }
      }
    },
    "custom": {
      "description": "A custom string",
      "pattern": "^[a-z]{4}$",
      "type": "string"
    }
  }
}
`
	assert.JSONEq(t, expected, string(swg), coloredJSONDiff(expected, string(swg)))
}

func TestGenerator_CapitalizeDefinitions(t *testing.T) {
	g := NewGenerator()
	g.CapitalizeDefinitions(true)
	g.SetPathItem(PathItemInfo{
		Method:   http.MethodPost,
		Path:     "/some",
		Response: new(testEmptyStruct),
	})

	expected := `{"swagger":"2.0","info":{"title":"","description":"","termsOfService":"","contact":{"name":""},"license":{"name":""},"version":""},"basePath":"/","schemes":["http","https"],"paths":{"/some":{"post":{"summary":"","description":"","responses":{"200":{"description":"OK","schema":{"$ref":"#/definitions/TestEmptyStruct"}}}}}},"definitions":{"TestEmptyStruct":{"type":"object"}}}`

	swg, err := g.GenDocument()
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(swg), coloredJSONDiff(expected, string(swg)))
}

func TestGenerator_SetPathItem_typeFile(t *testing.T) {
	type requestWithFileAndHeader struct {
		Upload       multipart.File        `file:"upload"`
		UploadHeader *multipart.FileHeader `file:"upload"`
	}
	type requestWithFile struct {
		Upload multipart.File `file:"upload"`
	}
	type requestWithHeader struct {
		UploadHeader *multipart.FileHeader `file:"upload"`
	}

	g := NewGenerator()
	g.SetPathItem(PathItemInfo{
		Method:  http.MethodPost,
		Path:    "/withFileAndHeader",
		Request: new(requestWithFileAndHeader),
	})
	g.SetPathItem(PathItemInfo{
		Method:  http.MethodPost,
		Path:    "/withFile",
		Request: new(requestWithFile),
	})
	g.SetPathItem(PathItemInfo{
		Method:  http.MethodPost,
		Path:    "/withHeader",
		Request: new(requestWithHeader),
	})

	expected := `{"swagger":"2.0","info":{"title":"","description":"","termsOfService":"","contact":{"name":""},"license":{"name":""},"version":""},"basePath":"/","schemes":["http","https"],"paths":{"/withFile":{"post":{"summary":"","description":"","parameters":[{"type":"file","name":"upload","in":"formData"}],"responses":{"204":{"description":"No Content"}}}},"/withFileAndHeader":{"post":{"summary":"","description":"","parameters":[{"type":"file","name":"upload","in":"formData"}],"responses":{"204":{"description":"No Content"}}}},"/withHeader":{"post":{"summary":"","description":"","parameters":[{"type":"file","name":"upload","in":"formData"}],"responses":{"204":{"description":"No Content"}}}}}}`

	swg, err := g.GenDocument()
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(swg), coloredJSONDiff(expected, string(swg)))
}

type CountryCode string

func (CountryCode) SwaggerDef() SwaggerData {
	def := SwaggerData{}
	def.Description = "Country Code"
	def.Example = "us"
	def.Type = "string"
	def.Pattern = "^[a-zA-Z]{2}$"
	return def
}

func TestGenerator_ParseParameters_namedSchemaParamItem(t *testing.T) {
	type Req struct {
		Codes []CountryCode `query:"countries" collectionFormat:"csv"`
	}

	g := NewGenerator()
	pathItem := PathItemInfo{
		Request: new(Req),
		Method:  http.MethodGet,
		Path:    "/some",
	}
	g.SetPathItem(pathItem)
	expected := `{"swagger":"2.0","info":{"title":"","description":"","termsOfService":"","contact":{"name":""},"license":{"name":""},"version":""},"basePath":"/","schemes":["http","https"],"paths":{"/some":{"get":{"summary":"","description":"","parameters":[{"type":"array","name":"countries","in":"query","items":{"type":"string","pattern":"^[a-zA-Z]{2}$"},"collectionFormat":"csv"}],"responses":{"204":{"description":"No Content"}}}}},"definitions":{"CountryCode":{"description":"Country Code","type":"string","pattern":"^[a-zA-Z]{2}$","example":"us"}}}`

	swg, err := g.GenDocument()
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(swg), coloredJSONDiff(expected, string(swg)))
}

func TestGenerator_ParseParameters(t *testing.T) {
	type Emb struct {
		P1 string `query:"p1"`
	}

	type Req struct {
		P0 int `path:"p0"`
		Emb
	}

	g := NewGenerator()

	name, params := g.ParseParameters(new(Req))
	assert.Equal(t, "Req", name)
	assert.Len(t, params, 2)
	assert.Equal(t, "p0", params[0].Name)
	assert.Equal(t, "integer", params[0].Type)
	assert.Equal(t, "p1", params[1].Name)
	assert.Equal(t, "string", params[1].Type)
}

func TestGenerator_SetPathItem_bodyMap(t *testing.T) {
	type (
		Value struct {
			S string `json:"s"`
		}
		Key string
		Map map[Key]Value
	)

	g := NewGenerator()
	g.AddPackagePrefix(true)
	obj := g.SetPathItem(PathItemInfo{
		Method:  http.MethodPost,
		Path:    "/",
		Request: new(Map),
	})

	assert.Len(t, obj.Parameters, 1)
	swg, err := g.GenDocument()
	assert.NoError(t, err)
	assert.Equal(t, `{"swagger":"2.0","info":{"title":"","description":"","termsOfService":"","contact":{"name":""},"license":{"name":""},"version":""},"basePath":"/","schemes":["http","https"],"paths":{"/":{"post":{"summary":"","description":"","parameters":[{"name":"body","in":"body","schema":{"$ref":"#/definitions/SwgenMap"},"required":true}],"responses":{"204":{"description":"No Content"}}}}},"definitions":{"SwgenMap":{"type":"object","additionalProperties":{"$ref":"#/definitions/SwgenValue"}},"SwgenValue":{"type":"object","properties":{"s":{"type":"string"}}}}}`, string(swg))
}
