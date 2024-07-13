package validation

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

// Add custom validators to handle validation scenarios not supported out of the box.
func configureCustomValidators() {
	V.RegisterValidation(ArchetypeValidator, func(fl validator.FieldLevel) bool {
		return len(archetypeRegex.FindAllString(fl.Field().String(), -1)) > 0
	})

	V.RegisterValidation(systemNameValidator, func(fl validator.FieldLevel) bool {
		return len(systemNameRegex.FindAllString(fl.Field().String(), -1)) > 0
	})

	V.RegisterValidation(systemVersionValidator, func(fl validator.FieldLevel) bool {
		return len(systemVersionRegex.FindAllString(fl.Field().String(), -1)) > 0
	})

	V.RegisterValidation(ygoCardIDsValidator, func(fl validator.FieldLevel) bool {
		cardIDs := fl.Field().Interface().(model.CardIDs)

		for _, cardID := range cardIDs {
			if len(cardIDRegex.FindAllString(cardID, -1)) == 0 {
				slog.Info("Deck Mascot ID not in proper format")
				return false
			}
		}

		return true
	})
}
