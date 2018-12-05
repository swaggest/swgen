package swgen_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/swgen"
)

type TestStruct1 struct {
	ID   uint
	Name string
}

type TestStruct2 struct {
	ID   uint
	Name string
}

func TestReflectTypeHash(t *testing.T) {
	var (
		ts1a, ts1b TestStruct1
		ts2        TestStruct2

		anon1a, anon1b struct {
			ID   uint
			Name string
		}

		anon2 = struct {
			ID   uint
			Name string
		}{}
	)

	if reflect.TypeOf(ts1a) != reflect.TypeOf(ts1b) {
		t.Error("Different reflect.Type on instances of the same named struct")
	}

	if swgen.ReflectTypeHash(reflect.TypeOf(ts1a)) == swgen.ReflectTypeHash(reflect.TypeOf(ts2)) {
		t.Error("Same reflect.Type on instances of different named structs:", swgen.ReflectTypeHash(reflect.TypeOf(ts1a)))
	}

	if reflect.TypeOf(anon1a) != reflect.TypeOf(anon1b) {
		t.Error("Different reflect.Type on instances of the same anonymous struct")
	}

	if swgen.ReflectTypeHash(reflect.TypeOf(anon1a)) != swgen.ReflectTypeHash(reflect.TypeOf(anon2)) {
		t.Error("Different reflect.Type on instances of the different anonymous structs with same fields")
	}
}

type (
	structWithEmbedded struct {
		B int `path:"b"`
		embedded
	}

	embedded struct {
		A int `json:"a"`
	}
)

func TestObjectHasXFields(t *testing.T) {
	assert.True(t, swgen.ObjectHasXFields(new(structWithEmbedded), "json"))
	assert.True(t, swgen.ObjectHasXFields(new(structWithEmbedded), "path"))
	assert.False(t, swgen.ObjectHasXFields(new(structWithEmbedded), "query"))
}
