package swgen

import "testing"

func TestSchemaFromCommonName(t *testing.T) {
	so := schemaFromCommonName(commonNameInteger)
	assertTrue(so.Type == "integer", t)
	assertTrue(so.Format == "int32", t)

	so = schemaFromCommonName("file")
	assertTrue(so.Type == "file", t)
	assertTrue(so.Format == "", t)
}
