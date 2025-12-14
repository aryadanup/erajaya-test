package utils

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Parameter string `json:"parameter"`
}

func ParseValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			field := fe.Field()
			tag := fe.Tag()

			var message string

			switch tag {
			case "required":
				message = field + " is required"
			case "min":
				message = field + " must be at least " + fe.Param()
			case "max":
				message = field + " must be at most " + fe.Param()
			case "email":
				message = field + " must be a valid email"
			default:
				message = field + " failed validation on tag '" + tag + "'"
			}

			errors = append(errors, ValidationError{
				Parameter: message,
			})
		}
	}

	return errors
}

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}

func NewValidator() *CustomValidator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return &CustomValidator{Validator: v}
}
