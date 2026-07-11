package validation

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	cModel "github.com/ygo-skc/skc-go/common/v2/model"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func Validate(tai model.TrafficData) *ValidationErrors {
	if err := V.Struct(tai); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			return HandleValidationErrors(ve)
		}
		slog.Error("Unexpected error while validating input", "err", err)
		return nil
	}
	return nil
}

func ValidateBatchCardIDs(bci cModel.BatchCardIDs) *ValidationErrors {
	if err := V.Struct(bci); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			return HandleValidationErrors(ve)
		}
		slog.Error("Unexpected error while validating input", "err", err)
		return nil
	}
	return nil
}
