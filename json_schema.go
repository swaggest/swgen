package swgen

import "strings"

type schemaBuilder struct {
	refs map[string]bool
	g    *Generator
}

// JsonSchema builds JSON Schema for Swagger Schema object
func (g *Generator) JsonSchema(s SchemaObj) (map[string]interface{}, error) {
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

		for ref, _ := range refs {
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

// ParamJsonSchema builds JSON Schema for Swagger Parameter object
func (g *Generator) ParamJsonSchema(p ParamObj) (map[string]interface{}, error) {
	if p.Schema != nil {
		return g.JsonSchema(*p.Schema)
	}

	p.Name = ""
	p.In = ""
	p.Required = false

	res, err := jsonRecode(p)
	return res, err
}
