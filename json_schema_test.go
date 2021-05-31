package swgen_test

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/swgen"
)

type foo struct {
	ID   int64  `path:"id" minimum:"1000"`
	Name string `json:"name" minLength:"10"`
	Bar  bar    `json:"bar"`
}

type bar struct {
	Yes bool `json:"yes"`
}

type baz struct {
	Name  string `json:"name"`
	Barrr bar    `json:"barrr"`
}

func TestGenerator_JSONSchemaWithCustomConfig(t *testing.T) {
	gen := swgen.NewGenerator()

	defs := make(map[string]map[string]interface{})

	obj := gen.SetPathItem(swgen.PathItemInfo{
		Request: new(foo),
		Method:  http.MethodPost,
		Path:    "/one/{id}/two",
	})

	cfg := swgen.JSONSchemaConfig{
		CollectDefinitions: defs,
		StripDefinitions:   true,
		DefinitionsPrefix:  "#/components/schema/",
	}

	g, err := gen.GetJSONSchemaRequestGroups(obj, cfg)
	assert.NoError(t, err)

	bodySchema, err := gen.GetJSONSchemaRequestBody(obj, cfg)
	assert.NoError(t, err)
	js, err := json.Marshal(bodySchema)
	assert.NoError(t, err)
	assert.Equal(t, `{"$ref":"#/components/schema/foo"}`, string(js))

	js, err = json.Marshal(g[`path`])
	assert.NoError(t, err)
	assert.Equal(t, `{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","required":["id"],"properties":{"id":{"format":"int64","minimum":1000,"type":"integer"}}}`, string(js))

	obj = gen.SetPathItem(swgen.PathItemInfo{
		Request: new(baz),
		Method:  http.MethodPost,
		Path:    "/three",
	})

	assert.NoError(t, err)

	_, err = gen.GetJSONSchemaRequestGroups(obj, cfg)
	assert.NoError(t, err)

	bodySchema, err = gen.GetJSONSchemaRequestBody(obj, cfg)
	assert.NoError(t, err)
	js, err = json.Marshal(bodySchema)
	assert.NoError(t, err)
	assert.Equal(t, `{"$ref":"#/components/schema/baz"}`, string(js))

	js, err = json.MarshalIndent(defs, "", " ")
	assert.NoError(t, err)
	assert.Equal(t, `{
 "bar": {
  "properties": {
   "yes": {
    "type": "boolean"
   }
  },
  "type": "object"
 },
 "baz": {
  "properties": {
   "barrr": {
    "$ref": "#/components/schema/bar"
   },
   "name": {
    "type": "string"
   }
  },
  "type": "object"
 },
 "foo": {
  "properties": {
   "bar": {
    "$ref": "#/components/schema/bar"
   },
   "name": {
    "minLength": 10,
    "type": "string"
   }
  },
  "type": "object"
 }
}`, string(js))
}

func TestGenerator_WalkJSONSchemaResponses(t *testing.T) {
	gen := swgen.NewGenerator()

	pathItem := swgen.PathItemInfo{
		Request:  new(foo),
		Method:   http.MethodPost,
		Path:     "/one/{id}/two",
		Response: new(baz),
	}
	pathItem.AddResponse(http.StatusNoContent, nil)
	gen.SetPathItem(pathItem)

	err := gen.WalkJSONSchemaResponses(func(path, method string, statusCode int, schema map[string]interface{}) {
		assert.Equal(t, "/one/{id}/two", path)
		assert.Equal(t, http.MethodPost, method)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, map[string]interface{}{
			"$schema": "http://json-schema.org/draft-04/schema#",
			"type":    "object",
			"properties": map[string]interface{}{
				"barrr": map[string]interface{}{"$ref": "#/definitions/bar"},
				"name":  map[string]interface{}{"type": "string"},
			},
			"definitions": map[string]map[string]interface{}{
				"bar": {
					"type": "object",
					"properties": map[string]interface{}{
						"yes": map[string]interface{}{"type": "boolean"},
					},
				},
			},
		}, schema)
	})
	assert.NoError(t, err)
}

func TestGenerator_WalkJSONSchemaRequestGroups(t *testing.T) {
	gen := swgen.NewGenerator()

	pathItem := swgen.PathItemInfo{
		Request:  new(foo),
		Method:   http.MethodPost,
		Path:     "/one/{id}/two",
		Response: new(baz),
	}
	pathItem.AddResponse(http.StatusNoContent, nil)
	gen.SetPathItem(pathItem)

	found := map[string]swgen.ObjectJSONSchema{}

	err := gen.WalkJSONSchemaRequestGroups(func(path, method, in string, schema swgen.ObjectJSONSchema) {
		assert.Equal(t, "/one/{id}/two", path)
		assert.Equal(t, http.MethodPost, method)
		_, exists := found[in]
		assert.False(t, exists)
		found[in] = schema
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(found))
	assert.Equal(t, []string{"id"}, found["path"].Required)
	assert.Equal(t, map[string]interface{}{"type": "integer", "format": "int64", "minimum": 1000.0}, found["path"].Properties["id"])
	_, hasBody := found["body"]
	assert.False(t, hasBody)
}

func TestGenerator_WalkJSONSchemaRequestBodies(t *testing.T) {
	gen := swgen.NewGenerator()

	pathItem := swgen.PathItemInfo{
		Request:  new(foo),
		Method:   http.MethodPost,
		Path:     "/one/{id}/two",
		Response: new(baz),
	}
	pathItem.AddResponse(http.StatusNoContent, nil)
	gen.SetPathItem(pathItem)

	var bodySchema map[string]interface{}

	err := gen.WalkJSONSchemaRequestBodies(func(path, method string, schema map[string]interface{}) {
		assert.Equal(t, "/one/{id}/two", path)
		assert.Equal(t, http.MethodPost, method)
		bodySchema = schema
	})
	assert.NoError(t, err)
	jsonBytes, err := json.Marshal(bodySchema)
	assert.NoError(t, err)
	assert.Equal(t, `{"definitions":{"bar":{"properties":{"yes":{"type":"boolean"}},"type":"object"}},"properties":{"bar":{"$ref":"#/definitions/bar"},"name":{"minLength":10,"type":"string"}},"type":"object"}`, string(jsonBytes))
}

func TestGenerator_ParamJSONSchema(t *testing.T) {
	gen := swgen.NewGenerator()

	type Req struct {
		F *multipart.FileHeader `file:"upload" description:"File Upload"`
	}

	pathItem := swgen.PathItemInfo{
		Request:  new(Req),
		Method:   http.MethodPost,
		Path:     "/one/{id}/two",
		Response: new(baz),
	}
	obj := gen.SetPathItem(pathItem)

	schema, err := gen.ParamJSONSchema(obj.Parameters[0])
	assert.NoError(t, err)
	jsonBytes, err := json.Marshal(schema)
	assert.NoError(t, err)
	assert.Equal(t, `{"description":"File Upload"}`, string(jsonBytes))
}
