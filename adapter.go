package swgen

import (
	"reflect"

	"github.com/swaggest/jsonschema-go"
)

var _ jsonschema.InterceptTypeFunc = JSONSchemaInterceptType

// JSONSchemaInterceptType is an adapter for github.com/swaggest/jsonschema-go.
func JSONSchemaInterceptType(v reflect.Value, s *jsonschema.Schema) (bool, error) {
	i := v.Interface()
	if def, ok := i.(SwaggerData); ok {
		LoadJSONSchemaFromSwgen(def, s)
		return true, nil
	}

	if def, ok := i.(SchemaDefinition); ok {
		d := def.SwaggerDef()
		LoadJSONSchemaFromSwgen(d, s)
	}

	return false, nil
}

// LoadJSONSchemaFromSwgen is an adapter for github.com/swaggest/jsonschema-go.
func LoadJSONSchemaFromSwgen(d SwaggerData, s *jsonschema.Schema) {
	if d.Type != "" {
		s.AddType(jsonschema.SimpleType(d.Type))
	}
	if d.Nullable {
		s.AddType(jsonschema.Null)
	}
	if d.UniqueItems {
		s.UniqueItems = &d.UniqueItems
	}
	if d.Title != "" {
		s.Title = &d.Title
	}
	if d.Description != "" {
		s.Description = &d.Description
	}
	if d.Format != "" {
		s.Format = &d.Format
	}
	if d.MinProperties != nil {
		s.MinProperties = *d.MinProperties
	}
	if d.MaxProperties != nil {
		s.MaxProperties = d.MaxProperties
	}
	if d.MinItems != nil {
		s.MinItems = *d.MinItems
	}
	if d.MaxItems != nil {
		s.MaxItems = d.MaxItems
	}
	if d.Minimum != nil {
		s.Minimum = d.Minimum
	}
	if d.Maximum != nil {
		s.Maximum = d.Maximum
	}
	if d.MinLength != nil {
		s.MinLength = *d.MinLength
	}
	if d.MaxLength != nil {
		s.MaxLength = d.MaxLength
	}
	if d.MultipleOf != 0 {
		s.MultipleOf = &d.MultipleOf
	}
	if d.Example != nil {
		s.Examples = append(s.Examples, d.Example)
	}
	if d.Default != nil {
		s.Default = &d.Default
	}
	if len(d.Enum.Enum) > 0 {
		s.Enum = d.Enum.Enum
		if len(d.EnumNames) > 0 {
			s.WithExtraPropertiesItem(jsonschema.XEnumNames, d.EnumNames)
		}
	}
	if d.Pattern != "" {
		s.Pattern = &d.Pattern
	}
}
