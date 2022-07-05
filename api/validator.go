package api

import (
	"log"
	"regexp"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var (
	v            = validator.New()
	translator   ut.Translator
	deckListName = regexp.MustCompile("^[a-zA-Z0-9 ]*$")
)

func init() {
	setupTranslators()
	v.RegisterValidation("decklistname", func(fl validator.FieldLevel) bool {
		return len(deckListName.FindAllString(fl.Field().String(), -1)) > 0
	})
}

func setupTranslators() {
	enTranslator := en.New()
	uni := ut.New(enTranslator, enTranslator)

	var found bool
	translator, found = uni.GetTranslator("en")
	if !found {
		log.Fatal("translator not found")
	}

	v.RegisterTranslation("decklistname", translator, func(ut ut.Translator) error {
		return ut.Add("decklistname", "Field {0} can only contain letters, numbers and spaces.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("decklistname", fe.Field())
		return t
	})

	v.RegisterTranslation("required", translator, func(ut ut.Translator) error {
		return ut.Add("required", "Field {0} is required.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	})

	v.RegisterTranslation("base64", translator, func(ut ut.Translator) error {
		return ut.Add("base64", "Field {0} needs to be properly encoded in base64.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("base64", fe.Field())
		return t
	})

	v.RegisterTranslation("url", translator, func(ut ut.Translator) error {
		return ut.Add("url", "Field {0} should be a proper url.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("url", fe.Field())
		return t
	})
}
