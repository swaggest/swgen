package swgen

import "strings"

type schemaBuilder struct {
	refs map[string]bool
	g    *Generator
}

// JSONSchema builds JSON Schema for Swagger Schema object
func (g *Generator) JSONSchema(s SchemaObj) (map[string]interface{}, error) {
	sb := &schemaBuilder{
		refs: make(map[string]bool),
		g:    g,
	}
	res, err := sb.jsonSchemaPlain(s)
	if err != nil {
		return nil, err
	}
	allDef := g.definitions.GenDefinitions()

	definitions := make(map[string]interface{})
	for len(sb.refs) > 0 {
		refs := sb.refs
		sb.refs = make(map[string]bool)

		for ref := range refs {
			ref := strings.TrimPrefix(ref, "#/definitions/")
			if _, ok := definitions[ref]; !ok {
				jsonSchema, err := sb.jsonSchemaPlain(allDef[ref])
				if err != nil {
					return nil, err
				}
				definitions[ref] = jsonSchema
			}
		}
	}

	if len(definitions) > 0 {
		res["definitions"] = definitions
	}
	return res, nil
}

func (sb *schemaBuilder) jsonSchemaPlain(s SchemaObj) (map[string]interface{}, error) {
	if s.Ref != "" {
		sb.refs[s.Ref] = true
		return map[string]interface{}{"$ref": s.Ref}, nil
	}
	res, err := jsonRecode(s)
	if err != nil {
		return nil, err
	}

	if s.Nullable && s.Type != "" {
		res["type"] = []interface{}{s.Type, "null"}
	}

	if s.Properties != nil && len(s.Properties) > 0 {
		properties := make(map[string]interface{}, len(s.Properties))
		for name, schema := range s.Properties {
			properties[name], err = sb.jsonSchemaPlain(schema)
			if err != nil {
				return nil, err
			}
		}
		res["properties"] = properties
	}

	if s.AdditionalProperties != nil {
		jsonSchema, err := sb.jsonSchemaPlain(*s.AdditionalProperties)
		if err != nil {
			return nil, err
		}
		res["additionalProperties"] = jsonSchema
	}

	if s.Items != nil {
		jsonSchema, err := sb.jsonSchemaPlain(*s.Items)
		if err != nil {
			return nil, err
		}
		res["items"] = jsonSchema
	}

	return res, nil
}

// ParamJSONSchema builds JSON Schema for Swagger Parameter object
func (g *Generator) ParamJSONSchema(p ParamObj) (map[string]interface{}, error) {
	if p.Schema != nil {
		return g.JSONSchema(*p.Schema)
	}

	p.Name = ""
	p.In = ""
	p.Required = false
	p.CollectionFormat = ""

	res, err := jsonRecode(p)
	return res, err
}

// ObjectJSONSchema is a simplified JSON Schema for object
type ObjectJSONSchema struct {
	ID         string                 `json:"id,omitempty"`
	Schema     string                 `json:"$schema,omitempty"`
	Type       string                 `json:"type"`
	Required   []string               `json:"required,omitempty"`
	Properties map[string]interface{} `json:"properties"`
}

// WalkJSONSchemaRequestGroups iterates over all request parameters grouped by path, method and in into an instance of JSON Schema
func (g *Generator) WalkJSONSchemaRequestGroups(function func(path, method, in string, schema ObjectJSONSchema)) {
	var err error
	for path, pi := range g.doc.Paths {
		for method, op := range pi.Map() {
			requestSchemas := map[string]ObjectJSONSchema{}
			for _, param := range op.Parameters {
				if _, ok := requestSchemas[param.In]; !ok {
					requestSchemas[param.In] = ObjectJSONSchema{
						Schema:     "http://json-schema.org/draft-04/schema#",
						Type:       "object",
						Required:   []string{},
						Properties: map[string]interface{}{},
					}
				}

				if param.Required {
					rs := requestSchemas[param.In]
					rs.Required = append(rs.Required, param.Name)
					requestSchemas[param.In] = rs
				}
				requestSchemas[param.In].Properties[param.Name], err = g.ParamJSONSchema(param)
				if err != nil {
					panic(err.Error())
				}
			}

			for in, schema := range requestSchemas {
				function(path, method, in, schema)
			}
		}
	}
}

// WalkJSONSchemaResponses iterates over all responses grouped by path, method and status code into an instance of JSON Schema
func (g *Generator) WalkJSONSchemaResponses(function func(path, method string, statusCode int, schema map[string]interface{})) {
	for path, pi := range g.doc.Paths {
		for method, op := range pi.Map() {
			for statusCode, resp := range op.Responses {
				schema, err := g.JSONSchema(*resp.Schema)
				if err != nil {
					panic(err.Error())
				}
				schema["$schema"] = "http://json-schema.org/draft-04/schema#"
				function(path, method, statusCode, schema)
			}
		}
	}

}
