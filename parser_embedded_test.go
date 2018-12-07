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
		embedded
	}

	structWithTaggedEmbedded struct {
		B        int `path:"b" json:"-"`
		embedded `json:"emb"`
	}

	structWithIgnoredEmbedded struct {
		B        int `path:"b" json:"-"`
		embedded `json:"-"`
	}

	embedded struct {
		A int `json:"a"`
	}
)

func TestGenerator_ParseDefinition_Embedded(t *testing.T) {
	b, err := json.Marshal(structWithTaggedEmbedded{B: 10, embedded: embedded{A: 20}})
	assert.NoError(t, err)
	assert.Equal(t, `{"emb":{"a":20}}`, string(b))

	b, err = json.Marshal(structWithEmbedded{B: 10, embedded: embedded{A: 20}})
	assert.NoError(t, err)
	assert.Equal(t, `{"a":20}`, string(b))

	b, err = json.Marshal(structWithIgnoredEmbedded{B: 10, embedded: embedded{A: 20}})
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
		Method:   http.MethodPost,
		Path:     "/structWithEmbedded",
		Response: new(structWithEmbedded),
	})
	g.SetPathItem(swgen.PathItemInfo{
		Method:   http.MethodPost,
		Path:     "/structWithIgnoredEmbedded",
		Response: new(structWithIgnoredEmbedded),
	})

	b, err = g.GenDocument()
	assert.NoError(t, err)
	assert.JSONEq(t,
		`{"swagger":"2.0","info":{"title":"","description":"","termsOfService":"","contact":{"name":""},"license":{"name":""},"version":""},"basePath":"/","schemes":["http","https"],"paths":{"/structWithEmbedded":{"post":{"summary":"","description":"","responses":{"200":{"description":"OK","schema":{"$ref":"#/definitions/structWithEmbedded"}}}}},"/structWithIgnoredEmbedded":{"post":{"summary":"","description":"","responses":{"200":{"description":"OK","schema":{"$ref":"#/definitions/structWithIgnoredEmbedded"}}}}},"/structWithTaggedEmbedded":{"post":{"summary":"","description":"","responses":{"200":{"description":"OK","schema":{"$ref":"#/definitions/structWithTaggedEmbedded"}}}}}},"definitions":{"embedded":{"type":"object","properties":{"a":{"type":"integer","format":"int32"}}},"structWithEmbedded":{"type":"object","properties":{"a":{"type":"integer","format":"int32"}}},"structWithIgnoredEmbedded":{"type":"object"},"structWithTaggedEmbedded":{"type":"object","properties":{"emb":{"$ref":"#/definitions/embedded"}}}}}`,
		string(b))

}
