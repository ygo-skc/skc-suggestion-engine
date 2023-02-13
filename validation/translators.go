package validation

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

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
