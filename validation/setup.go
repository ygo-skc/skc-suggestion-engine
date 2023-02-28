package validation

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

func HandleValidationErrors(err error) string {
	errMessages := []string{}
	for _, e := range err.(validator.ValidationErrors) {
		errMessages = append(errMessages, e.Translate(Translator))
	}

	message := strings.Join(errMessages, " ")
	log.Printf("There were %d errors while validating input. Errors: %s", len(errMessages), message)

	return message
}
