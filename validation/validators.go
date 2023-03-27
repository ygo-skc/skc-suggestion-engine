package validation

import (
	"log"

	"github.com/go-playground/validator/v10"
)

// Add custom validators to handle validation scenarios not supported out of the box.
func configureCustomValidators() {
	V.RegisterValidation(deckListNameValidator, func(fl validator.FieldLevel) bool {
		return len(deckListNameRegex.FindAllString(fl.Field().String(), -1)) > 0
	})

	V.RegisterValidation(deckMascotsValidator, func(fl validator.FieldLevel) bool {
		mascots := fl.Field().Interface().([]string)

		for ind, mascot := range mascots {
			if ind == 3 { // size constraint fails
				log.Println("Deck Mascot array failed size constraint.")
				return false
			} else if len(cardIDRegex.FindAllString(mascot, -1)) == 0 { // regex constraint
				log.Println("Deck Mascot ID not in proper format.")
				return false
			}
		}

		return true
	})

	V.RegisterValidation(systemNameValidator, func(fl validator.FieldLevel) bool {
		return len(systemNameRegex.FindAllString(fl.Field().String(), -1)) > 0
	})

	V.RegisterValidation(systemVersionValidator, func(fl validator.FieldLevel) bool {
		return len(systemVersionRegex.FindAllString(fl.Field().String(), -1)) > 0
	})
}
