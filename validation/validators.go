package validation

import (
	"github.com/go-playground/validator/v10"
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
}
