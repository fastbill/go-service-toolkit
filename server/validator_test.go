package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Foo string `validate:"required"`
}

func TestNew(t *testing.T) {
	v := NewValidator()
	assert.NotNil(t, v)
	assert.NotNil(t, v.validator)
}

func TestValidate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		v := NewValidator()
		input := testStruct{"foo"}
		err := v.Validate(input)
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		v := NewValidator()
		input := testStruct{""}
		err := v.Validate(input)
		if assert.Error(t, err) {
			assert.Equal(t, "Key: 'testStruct.Foo' Error:Field validation for 'Foo' failed on the 'required' tag", err.Error())
		}
	})
}
