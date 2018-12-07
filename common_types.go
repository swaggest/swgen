package swgen

type commonName string

const (
	// CommonNameInteger data type is integer, format int32 (signed 32 bits)
	commonNameInteger commonName = "integer"
	// CommonNameLong data type is integer, format int64 (signed 64 bits)
	commonNameLong commonName = "long"
	// CommonNameFloat data type is number, format float
	commonNameFloat commonName = "float"
	// CommonNameDouble data type is number, format double
	commonNameDouble commonName = "double"
	// CommonNameString data type is string
	commonNameString commonName = "string"
	// CommonNameByte data type is string, format byte (base64 encoded characters)
	commonNameByte commonName = "byte"
	// CommonNameBinary data type is string, format binary (any sequence of octets)
	commonNameBinary commonName = "binary"
	// CommonNameBoolean data type is boolean
	commonNameBoolean commonName = "boolean"
	// CommonNameDate data type is string, format date (As defined by full-date - RFC3339)
	commonNameDate commonName = "date"
	// CommonNameDateTime data type is string, format date-time (As defined by date-time - RFC3339)
	commonNameDateTime commonName = "dateTime"
	// CommonNamePassword data type is string, format password
	commonNamePassword commonName = "password"
)

type typeFormat struct {
	Type   string
	Format string
}

var commonNamesMap = map[commonName]typeFormat{
	commonNameInteger:  {"integer", "int32"},
	commonNameLong:     {"integer", "int64"},
	commonNameFloat:    {"number", "float"},
	commonNameDouble:   {"number", "double"},
	commonNameString:   {"string", ""},
	commonNameByte:     {"string", "byte"},
	commonNameBinary:   {"string", "binary"},
	commonNameBoolean:  {"boolean", ""},
	commonNameDate:     {"string", "date"},
	commonNameDateTime: {"string", "date-time"},
	commonNamePassword: {"string", "password"},
}

func isCommonName(typeName string) (ok bool) {
	_, ok = commonNamesMap[commonName(typeName)]
	return
}

// schemaFromCommonName create SchemaObj from common name of data types
// supported types: https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types
func schemaFromCommonName(name commonName) SchemaObj {
	data, ok := commonNamesMap[name]
	if ok {
		return SchemaObj{shared: shared{
			Type:   data.Type,
			Format: data.Format,
		}}
	}

	return SchemaObj{shared: shared{
		Type: string(name),
	}}
}
