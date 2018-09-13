package swgen

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

type Person struct {
	Name       *PersonName `json:"name"`
	SecondName PersonName  `json:"second_name"`
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
	if _, err := gen.ParseDefinition(ts); err != nil {
		t.Fatalf("%v", err)
	}

	if len(gen.definitions) == 0 {
		t.Fatalf("len of gen.definitions must be greater than 0")
	}

	gen.ResetDefinitions()
	if len(gen.definitions) != 0 {
		t.Fatalf("len of gen.definitions must be equal to 0")
	}
}

func TestParseDefinition(t *testing.T) {
	ts := &Person{}
	if _, err := NewGenerator().ParseDefinition(ts); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestParseDefinitionEmptyInterface(t *testing.T) {
	var ts interface{}
	gen := NewGenerator()
	if _, err := gen.ParseDefinition(&ts); err != nil {
		t.Fatalf("%v", err)
	}
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

	if _, err := NewGenerator().ParseDefinition(&ts); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestParseDefinitionWithEmbeddedStruct(t *testing.T) {
	ts := &Employee{}
	tt := reflect.TypeOf(ts)

	gen := NewGenerator()
	if _, err := gen.ParseDefinition(ts); err != nil {
		t.Fatalf("%v", err)
	}

	if typeDef, found := gen.getDefinition(tt); found == false {
		t.Fatal("No definition for", tt)
	} else {
		propertiesCount := len(typeDef.Properties)
		expectedPropertiesCount := 9
		if propertiesCount != expectedPropertiesCount {
			t.Fatalf("Expected %d properties, got %d : %#v", expectedPropertiesCount, propertiesCount, typeDef.Properties)
		}
	}
}

func TestParseDefinitionWithEmbeddedInterface(t *testing.T) {
	p := &Project{Manager: new(Employee)}
	tt := reflect.TypeOf(p)

	gen := NewGenerator()
	if _, err := gen.ParseDefinition(p); err != nil {
		t.Fatalf("%v", err)
	}

	if typeDef, found := gen.getDefinition(tt); found == false {
		t.Fatal("No definition for", tt)
	} else {
		if typeDef.Properties["manager"].Ref != "#/definitions/Employee" {
			t.Fatalf("'manager' field was not parsed correctly.")
		}
	}
}

func TestParseDefinitionString(t *testing.T) {
	typeDef, err := NewGenerator().ParseDefinition("string")
	name := typeDef.TypeName
	if err != nil {
		t.Fatalf("Error parsing string: %+v", err)
	}
	if name != "string" {
		t.Fatalf("Wrong type name. Expect %q, got %q", "string", name)
	}
}

func TestParseDefinitionArray(t *testing.T) {
	type Names []string
	typeDef, err := NewGenerator().ParseDefinition(Names{})
	if err != nil {
		t.Fatalf("Error while parsing array of string: %v", err)
	}

	if typeDef.TypeName != "Names" {
		t.Fatalf("Wrong type name. Expected: Names, Obtained: %v", typeDef.TypeName)
	}

	// re-parse with pointer input
	// should get from definition list
	_, err = NewGenerator().ParseDefinition(&Names{})
	if err != nil {
		t.Fatalf("Error while parsing array of string: %v", err)
	}

	// try to parse a named map
	type MapList map[string]string
	_, err = NewGenerator().ParseDefinition(&MapList{})
	if err != nil {
		t.Fatalf("Error while parsing map string to string: %v", err)
	}

	// named array of object
	type Person struct{}
	type Persons []*Person
	_, err = NewGenerator().ParseDefinition(&Persons{})
	if err != nil {
		t.Fatalf("Error while parsing array of object: %v", err)
	}
}

func TestParseParameter(t *testing.T) {
	p := &PreferredWarehouseRequest{}
	name, params, err := NewGenerator().ParseParameter(p)

	if err != nil {
		t.Fatalf("error %v", err)
	}

	if name != "PreferredWarehouseRequest" {
		t.Fatalf("name of parameter is %s, expected is PreferredWarehouseRequest", name)
	}

	if len(params) != 2 {
		t.Fatalf("number of parameter should be 2")
	}
}

func TestParseParameterError(t *testing.T) {
	_, _, err := NewGenerator().ParseParameter(true)
	if err == nil {
		t.Fatalf("it should return error")
	}
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
		_, err := gen.SetPathItem(info)
		if err != nil {
			t.Fatalf("error %v", err)
		}
	}

	if len(gen.paths) == 0 {
		t.Fatalf("len of gen.paths must be greater than 0")
	}

	gen.ResetPaths()
	if len(gen.paths) != 0 {
		t.Fatalf("len of gen.paths must be equal to 0")
	}

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
	ID int    `query:"id"`
	C  custom `query:"c"`
}

func TestSwaggerDef(t *testing.T) {
	gen := NewGenerator()

	_, err := gen.SetPathItem(PathItemInfo{
		Method:  http.MethodPost,
		Path:    "/bla",
		Request: new(bodyWithCustom),
	})
	if err != nil {
		t.Fatalf("error while adding POST handler: %v", err)
	}

	_, err = gen.SetPathItem(PathItemInfo{
		Method:  http.MethodGet,
		Path:    "/bla",
		Request: new(paramsWithCustom),
	})
	if err != nil {
		t.Fatalf("error while adding GET handler: %v", err)
	}

	swg, err := gen.GenDocument()
	if err != nil {
		t.Fatalf("error while generating swagger doc: %v", err)
	}
	expected := []byte(`
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
          "200": {
            "schema": {
              "type": "null"
            }
          }
        }
      },
      "post": {
        "summary": "",
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
          "200": {
            "schema": {
              "type": "null"
            }
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
`)

	assertEqualJSON(swg, expected, t)
}
