package validation

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// overrides a validators default message and sets it's error message
func overrideTranslation(tag string, errorMessage string) func(ut.Translator) error {
	return func(ut ut.Translator) error {
		return ut.Add(tag, errorMessage, true) // see universal-translator for details
	}
}

// uses registered validator and its message and transforms the message using the field value that triggered the validation error
func createTranslation(tag string) func(ut.Translator, validator.FieldError) string {
	return func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field())
		return t
	}
}

// simplifies registration of new validation and error message
func registerTranslation(tag string, description string) {
	V.RegisterTranslation(tag, Translator, overrideTranslation(tag, description), createTranslation(tag))
}

// Add translations for errors so messages are more informative.
func configureTranslations() {
	registerTranslation(deckListNameValidator, "{0} can only contain letters, numbers, spaces and the following special characters: @-!. It must also be at least 3 characters long.")
	registerTranslation(deckMascotsValidator, "{0} failed one of the following constraints: can only contain the 8 digit ID of the card, must contain no more than 3 mascots.")
	registerTranslation(requiredValidator, "{0} is required.")
	registerTranslation(base64Validator, "{0} needs to be properly encoded in base64.")
	registerTranslation(urlValidator, "{0} should be a proper url.")
	registerTranslation(systemNameValidator, "{0} can only contain letters, numbers, spaces and the special character -.")
	registerTranslation(systemVersionValidator, "{0} should use major.minor.patch (Semantic Versioning) format.")
	registerTranslation(ipv4Validator, "{0} should use ipv4 format.")
}
