package util

import (
	"log"
	"regexp"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var (
	V            = validator.New()
	Translator   ut.Translator
	deckListName = regexp.MustCompile("^[a-zA-Z0-9 ]*$")
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
		return len(deckListName.FindAllString(fl.Field().String(), -1)) > 0
	})
}

// Add translations for errors so messages are more informative.
func configureTranslations() {
	V.RegisterTranslation("decklistname", Translator, func(ut ut.Translator) error {
		return ut.Add("decklistname", "Field {0} can only contain letters, numbers and spaces.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("decklistname", fe.Field())
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
