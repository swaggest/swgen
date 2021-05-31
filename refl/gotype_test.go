package refl_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/refl"
	fancypath "github.com/swaggest/swgen/internal/Fancy-Path"
	"github.com/swaggest/swgen/internal/sample"
)

func TestGoType(t *testing.T) {
	assert.Equal(
		t,
		refl.TypeString("github.com/swaggest/swgen/internal/sample.TestSampleStruct"),
		refl.GoType(reflect.TypeOf(sample.TestSampleStruct{})),
	)
	assert.Equal(
		t,
		refl.TypeString("*github.com/swaggest/swgen/internal/sample.TestSampleStruct"),
		refl.GoType(reflect.TypeOf(new(sample.TestSampleStruct))),
	)
	assert.Equal(
		t,
		refl.TypeString("*github.com/swaggest/swgen/internal/sample.TestSampleStruct"),
		refl.GoType(reflect.TypeOf(new(sample.TestSampleStruct))),
	)
	assert.Equal(
		t,
		refl.TypeString("*github.com/swaggest/swgen/internal/Fancy-Path::fancypath.Sample"),
		refl.GoType(reflect.TypeOf(new(fancypath.Sample))),
	)
	assert.Equal(
		t,
		refl.TypeString("*[]map[*github.com/swaggest/swgen/internal/Fancy-Path::fancypath.Sample]github.com/swaggest/swgen/internal/Fancy-Path::fancypath.Sample"),
		refl.GoType(reflect.TypeOf(new([]map[*fancypath.Sample]fancypath.Sample))),
	)
}
