package swgen

import (
	"fmt"
	"reflect"
)

// ReflectTypeHash returns private (unexported) `hash` field of the Golang internal reflect.rtype struct for a given reflect.Type
// This hash is used to (quasi-)uniquely identify a reflect.Type value
func ReflectTypeHash(t reflect.Type) uint32 {
	return uint32(reflect.Indirect(reflect.ValueOf(t)).FieldByName("hash").Uint())
}

// ReflectTypeReliableName returns real name of given reflect.Type, if it is non-empty, or auto-generates "anon_*"]
// name for anonymous structs
func ReflectTypeReliableName(t reflect.Type) string {
	if def, ok := reflect.Zero(t).Interface().(SchemaDefinition); ok {
		typeDef := def.SwaggerDef()
		if typeDef.TypeName != "" {
			return typeDef.TypeName
		}
	}
	if t.Name() != "" {
		return t.Name()
	}
	return fmt.Sprintf("anon_%08x", ReflectTypeHash(t))
}

// ObjectHasXFields checks if the structure has fields with tag name
func ObjectHasXFields(i interface{}, tagname string) bool {
	if i == nil {
		return false
	}
	t := reflect.TypeOf(i)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get(tagname); tag != "" && tag != "-" {
			return true
		}
	}
	return false
}

func IsSlice(i interface{}) bool {
	t := reflect.TypeOf(i)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		return true
	}
	return false
}
