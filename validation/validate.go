package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func Validate(tai model.TrafficData) *ValidationErrors {
	if err := V.Struct(tai); err != nil {
		return HandleValidationErrors(err.(validator.ValidationErrors))
	} else {
		return nil
	}
}
