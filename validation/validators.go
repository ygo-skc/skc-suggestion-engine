package validation

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	cModel "github.com/ygo-skc/skc-go/common/v3/model"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

// Add custom validators to handle validation scenarios not supported out of the box.
func configureCustomValidators() {
	V.RegisterValidation(ArchetypeValidator, func(fl validator.FieldLevel) bool {
		return archetypeRegex.MatchString(fl.Field().String())
	})

	V.RegisterValidation(systemNameValidator, func(fl validator.FieldLevel) bool {
		return systemNameRegex.MatchString(fl.Field().String())
	})

	V.RegisterValidation(systemVersionValidator, func(fl validator.FieldLevel) bool {
		return systemVersionRegex.MatchString(fl.Field().String())
	})

	V.RegisterValidation(ygoCardIDsValidator, func(fl validator.FieldLevel) bool {
		cardIDs := fl.Field().Interface().(cModel.CardIDs)

		for _, cardID := range cardIDs {
			if !cardIDRegex.MatchString(cardID) {
				slog.Info("Deck Mascot ID not in proper format")
				return false
			}
		}

		return true
	})

	V.RegisterValidation(trendingResourceValidator, func(fl validator.FieldLevel) bool {
		return fl.Field().String() == string(model.CardResource) || fl.Field().String() == string(model.ProductResource)
	})
}
