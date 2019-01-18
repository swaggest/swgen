package swgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathItemHasMethod(t *testing.T) {
	item := PathItem{}
	item.Get = &OperationObj{}

	assert.True(t, item.HasMethod("GET"))
	assert.False(t, item.HasMethod("POST"))
	assert.False(t, item.HasMethod("PUT"))
	assert.False(t, item.HasMethod("HEAD"))
	assert.False(t, item.HasMethod("DELETE"))
	assert.False(t, item.HasMethod("OPTIONS"))
	assert.False(t, item.HasMethod("PATCH"))
	assert.False(t, item.HasMethod(""))
}

func TestAdditionalDataJSONMarshal(t *testing.T) {
	// empty object
	obj := additionalData{}
	_, err := obj.marshalJSONWithStruct(nil)
	assert.NoError(t, err)

	obj.AddExtendedField("x-custom-field", 1)
	data, err := obj.marshalJSONWithStruct(struct{}{})
	assert.NoError(t, err)
	assert.Equal(t, `{"x-custom-field":1}`, string(data))
}
