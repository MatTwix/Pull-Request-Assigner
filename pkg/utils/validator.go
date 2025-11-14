package utils

import "github.com/go-playground/validator"

var validate *validator.Validate

func InitValidator() {
	validate = validator.New()
}

func ValidateStruct(s any) error {
	return validate.Struct(s)
}
