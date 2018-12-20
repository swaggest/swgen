package swgen_test

import (
	"encoding/json"
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

	js, err := json.Marshal(g[`path`])
	assert.NoError(t, err)
	assert.Equal(t, `{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","required":["id"],"properties":{"id":{"format":"int64","minimum":1000,"type":"integer"}}}`, string(js))

	js, err = json.Marshal(g[`body`].Properties[`body`])
	assert.NoError(t, err)
	assert.Equal(t, `{"$ref":"#/components/schema/foo"}`, string(js))

	obj = gen.SetPathItem(swgen.PathItemInfo{
		Request: new(baz),
		Method:  http.MethodPost,
		Path:    "/three",
	})
	assert.NoError(t, err)

	g, err = gen.GetJSONSchemaRequestGroups(obj, cfg)
	assert.NoError(t, err)

	js, err = json.Marshal(g[`body`].Properties[`body`])
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
