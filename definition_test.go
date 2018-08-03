package swgen

import (
	"testing"
)

func TestDefinition(t *testing.T) {
	var obj = SwaggerData{}
	obj.Type = "integer"
	obj.Format = "int64"
	obj.TypeName = "MyName"

	typeDef := obj.SwaggerDef()
	assertTrue(typeDef.TypeName == "MyName", t)
}
