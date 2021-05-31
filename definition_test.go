package swgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefinition(t *testing.T) {
	obj := SwaggerData{}
	obj.Type = "integer"
	obj.Format = "int64"
	obj.TypeName = "MyName"

	typeDef := obj.SwaggerDef()
	assert.Equal(t, "MyName", typeDef.TypeName)
}
