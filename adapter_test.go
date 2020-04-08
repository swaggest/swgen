package swgen_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swaggest/assertjson"
	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/swgen"
)

// ISOWeek is a week identifier.
type ISOWeek string

// SwaggerDef returns swagger definition.
func (ISOWeek) SwaggerDef() swgen.SwaggerData {
	s := swgen.SwaggerData{}

	s.Description = "ISO Week"
	s.Example = "2006-W43"
	s.Type = "string"
	s.Pattern = `^[0-9]{4}-W(0[1-9]|[1-4][0-9]|5[0-2])$`

	return s
}

type UUID [16]byte

func TestInterceptType(t *testing.T) {
	reflector := jsonschema.Reflector{}
	reflector.DefaultOptions = append(reflector.DefaultOptions, jsonschema.InterceptType(swgen.JSONSchemaInterceptType))

	// Add custom type mappings
	uuidDef := swgen.SwaggerData{}
	uuidDef.Type = "string"
	uuidDef.Format = "uuid"
	uuidDef.Example = "248df4b7-aa70-47b8-a036-33ac447e668d"

	reflector.AddTypeMapping(UUID{}, uuidDef)

	type MyStruct struct {
		UUID    UUID    `json:"uuid"`
		ISOWeek ISOWeek `json:"iso_week"`
	}

	schema, err := reflector.Reflect(MyStruct{})
	require.NoError(t, err)

	js, err := json.MarshalIndent(schema, "", " ")
	require.NoError(t, err)
	assertjson.Equal(t, []byte(`{
	 "definitions": {
	  "SwjschemaTestISOWeek": {
	   "description": "ISO Week",
	   "examples": [
		"2006-W43"
	   ],
	   "pattern": "^[0-9]{4}-W(0[1-9]|[1-4][0-9]|5[0-2])$",
	   "type": "string"
	  },
	  "SwjschemaTestUUID": {
	   "examples": [
		"248df4b7-aa70-47b8-a036-33ac447e668d"
	   ],
	   "type": "string",
	   "format": "uuid"
	  }
	 },
	 "properties": {
	  "iso_week": {
	   "$ref": "#/definitions/SwjschemaTestISOWeek"
	  },
	  "uuid": {
	   "$ref": "#/definitions/SwjschemaTestUUID"
	  }
	 },
	 "type": "object"
	}`), js, string(js))
}
