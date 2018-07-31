package swgen

import (
	"testing"
)

func TestDefinition(t *testing.T) {
	var obj SchemaDefinition
	obj = SchemaObj{Type: "integer", Format: "int64", TypeName: "MyName"}

	typeDef := obj.SwaggerDef()
	assertTrue(typeDef.TypeName == "MyName", t)
}
