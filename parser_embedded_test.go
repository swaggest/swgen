package swgen_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/swgen"
)

type (
	structWithEmbedded struct {
		B int `path:"b" json:"-"`
		Embedded
	}

	structWithEmbeddedInBody struct {
		B        int `path:"b" json:"-"`
		Embedded `in:"body"`
	}

	structWithTaggedEmbedded struct {
		B        int `path:"b" json:"-"`
		Embedded `json:"emb"`
	}

	structWithIgnoredEmbedded struct {
		B        int `path:"b" json:"-"`
		Embedded `json:"-"`
	}

	Embedded struct {
		A int `json:"a" query:"a"`
	}
)

func TestGenerator_ParseDefinition_Embedded(t *testing.T) {
	b, err := json.Marshal(structWithTaggedEmbedded{B: 10, Embedded: Embedded{A: 20}})
	assert.NoError(t, err)
	assert.Equal(t, `{"emb":{"a":20}}`, string(b))

	b, err = json.Marshal(structWithEmbedded{B: 10, Embedded: Embedded{A: 20}})
	assert.NoError(t, err)
	assert.Equal(t, `{"a":20}`, string(b))

	b, err = json.Marshal(structWithIgnoredEmbedded{B: 10, Embedded: Embedded{A: 20}})
	assert.NoError(t, err)
	assert.Equal(t, `{}`, string(b))

	g := swgen.NewGenerator()
	//g.IndentJSON(true)
	g.SetPathItem(swgen.PathItemInfo{
		Method:   http.MethodPost,
		Path:     "/structWithTaggedEmbedded",
		Response: new(structWithTaggedEmbedded),
	})
	g.SetPathItem(swgen.PathItemInfo{
		Method:  http.MethodPost,
		Path:    "/structWithEmbeddedInBody",
		Request: new(structWithEmbeddedInBody),
	})
	g.SetPathItem(swgen.PathItemInfo{
		Method:   http.MethodPost,
		Path:     "/structWithEmbedded",
		Request:  new(structWithEmbedded),
		Response: new(structWithEmbedded),
	})

	g.SetPathItem(swgen.PathItemInfo{
		Method:   http.MethodPost,
		Path:     "/structWithIgnoredEmbedded",
		Response: new(structWithIgnoredEmbedded),
	})

	b, err = g.IndentJSON(true).GenDocument()
	assert.NoError(t, err)
	assert.JSONEq(t,
		`{
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
    "/structWithEmbedded": {
      "post": {
        "summary": "",
        "description": "",
        "parameters": [
          {
            "type": "integer",
            "format": "int32",
            "name": "b",
            "in": "path",
            "required": true
          },
          {
            "type": "integer",
            "format": "int32",
            "name": "a",
            "in": "query"
          },
          {
            "name": "body",
            "in": "body",
            "schema": {
              "$ref": "#/definitions/structWithEmbedded"
            },
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "$ref": "#/definitions/structWithEmbedded"
            }
          }
        }
      }
    },
    "/structWithEmbeddedInBody": {
      "post": {
        "summary": "",
        "description": "",
        "parameters": [
          {
            "type": "integer",
            "format": "int32",
            "name": "b",
            "in": "path",
            "required": true
          },
          {
            "name": "body",
            "in": "body",
            "schema": {
              "$ref": "#/definitions/structWithEmbeddedInBody"
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
    },
    "/structWithIgnoredEmbedded": {
      "post": {
        "summary": "",
        "description": "",
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "$ref": "#/definitions/structWithIgnoredEmbedded"
            }
          }
        }
      }
    },
    "/structWithTaggedEmbedded": {
      "post": {
        "summary": "",
        "description": "",
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "$ref": "#/definitions/structWithTaggedEmbedded"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "Embedded": {
      "type": "object",
      "properties": {
        "a": {
          "type": "integer",
          "format": "int32"
        }
      }
    },
    "structWithEmbedded": {
      "type": "object",
      "properties": {
        "a": {
          "type": "integer",
          "format": "int32"
        }
      }
    },
    "structWithEmbeddedInBody": {
      "type": "object",
      "properties": {
        "a": {
          "type": "integer",
          "format": "int32"
        }
      }
    },
    "structWithIgnoredEmbedded": {
      "type": "object"
    },
    "structWithTaggedEmbedded": {
      "type": "object",
      "properties": {
        "emb": {
          "$ref": "#/definitions/Embedded"
        }
      }
    }
  }
}`,
		string(b))

}
