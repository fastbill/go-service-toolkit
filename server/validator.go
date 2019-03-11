package server

import (
	goValidator "github.com/go-playground/validator"
)

// Validator points to 3rd party validator package (library) which actually does the real validation
type Validator struct {
	validator *goValidator.Validate
}

// NewValidator creates new instance of the go-playground/validator
func NewValidator() *Validator {
	return &Validator{
		validator: goValidator.New(),
	}
}

// Validate performs validation of any data type which is mapped according to rules of the 3rd party validation library
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}
