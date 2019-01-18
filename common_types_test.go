package swgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaFromCommonName(t *testing.T) {
	so := schemaFromCommonName(commonNameInteger)
	assert.Equal(t, "integer", so.Type)
	assert.Equal(t, "int32", so.Format)

	so = schemaFromCommonName("file")
	assert.Equal(t, "file", so.Type)
	assert.Equal(t, "", so.Format)
}
