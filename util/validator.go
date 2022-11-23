package util

import (
	"log"
	"regexp"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var (
	V                  = validator.New()
	Translator         ut.Translator
	deckListNameRegex  = regexp.MustCompile("^[a-zA-Z0-9 !-@]*$")
	cardIDRegex        = regexp.MustCompile("[0-9]{8}")
	systemNameRegex    = regexp.MustCompile("^[a-zA-Z0-9 -]*$")
	systemVersionRegex = regexp.MustCompile(`^([1-9]\d*|0)(\.(([1-9]\d*)|0)){2,3}$`)
)

func init() {
	enTranslator := en.New()
	uni := ut.New(enTranslator, enTranslator)

	var found bool
	Translator, found = uni.GetTranslator("en")
	if !found {
		log.Fatal("translator not found")
	}

	configureTranslations()
	configureCustomValidators()
}

// Add custom validators to handle validation scenarios not supported out of the box.
func configureCustomValidators() {
	V.RegisterValidation("decklistname", func(fl validator.FieldLevel) bool {
		return len(deckListNameRegex.FindAllString(fl.Field().String(), -1)) > 0
	})

	V.RegisterValidation("deckmascots", func(fl validator.FieldLevel) bool {
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

	V.RegisterValidation("systemname", func(fl validator.FieldLevel) bool {
		return len(systemNameRegex.FindAllString(fl.Field().String(), -1)) > 0
	})

	V.RegisterValidation("systemversion", func(fl validator.FieldLevel) bool {
		return len(systemVersionRegex.FindAllString(fl.Field().String(), -1)) > 0
	})
}

// Add translations for errors so messages are more informative.
func configureTranslations() {
	V.RegisterTranslation("decklistname", Translator, func(ut ut.Translator) error {
		return ut.Add("decklistname", "Field {0} can only contain letters, numbers, spaces and the following special characters: @-!.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("decklistname", fe.Field())
		return t
	})

	V.RegisterTranslation("deckmascots", Translator, func(ut ut.Translator) error {
		return ut.Add("deckmascots", "Field {0} failed one of the following constraints: can only contain the 8 digit ID of the card, must contain no more than 3 mascots.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("deckmascots", fe.Field())
		return t
	})

	V.RegisterTranslation("required", Translator, func(ut ut.Translator) error {
		return ut.Add("required", "Field {0} is required.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	})

	V.RegisterTranslation("base64", Translator, func(ut ut.Translator) error {
		return ut.Add("base64", "Field {0} needs to be properly encoded in base64.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("base64", fe.Field())
		return t
	})

	V.RegisterTranslation("url", Translator, func(ut ut.Translator) error {
		return ut.Add("url", "Field {0} should be a proper url.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("url", fe.Field())
		return t
	})
}

func HandleValidationErrors(err error) string {
	errMessages := []string{}
	for _, e := range err.(validator.ValidationErrors) {
		errMessages = append(errMessages, e.Translate(Translator))
	}

	message := strings.Join(errMessages, " ")
	log.Printf("There were %d errors while validating input. Errors: %s", len(errMessages), message)

	return message
}
