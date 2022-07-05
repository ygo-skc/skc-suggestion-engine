package api

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	v            = validator.New()
	deckListName = regexp.MustCompile("^[a-zA-Z0-9 ]*$")
)

func init() {
	v.RegisterValidation("decklistname", func(fl validator.FieldLevel) bool {
		return len(deckListName.FindAllString(fl.Field().String(), -1)) > 0
	})
}
