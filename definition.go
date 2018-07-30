package swgen

// Definition is a helper that implements interface SchemaDefinition
type Definition struct {
	SchemaObj
}

// SwaggerSchema return type name and definition that was set
func (s Definition) SwgenDefinition() (typeName string, typeDef SchemaObj, err error) {
	typeName = s.TypeName
	typeDef = s.SchemaObj
	return
}
