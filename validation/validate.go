package validation

import (
	"github.com/go-playground/validator/v10"
	cModel "github.com/ygo-skc/skc-go/common/model"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func Validate(tai model.TrafficData) *ValidationErrors {
	if err := V.Struct(tai); err != nil {
		return HandleValidationErrors(err.(validator.ValidationErrors))
	} else {
		return nil
	}
}

func ValidateBatchCardIDs(bci cModel.BatchCardIDs) *ValidationErrors {
	if err := V.Struct(bci); err != nil {
		return HandleValidationErrors(err.(validator.ValidationErrors))
	} else {
		return nil
	}
}
