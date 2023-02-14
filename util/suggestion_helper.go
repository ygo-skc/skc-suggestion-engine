package util

import (
	"strings"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

// looks for a self reference, if a self reference is found it is removed from original slice
// this method returns true if a self reference is found
func RemoveSelfReference(self string, cr *[]model.CardReference) bool {
	hasSelfRef := false

	if cr != nil {
		x := 0
		for _, ref := range *cr {
			if ref.Card.CardName != self {
				(*cr)[x] = ref
				x++
			} else {
				hasSelfRef = true
			}
		}

		*cr = (*cr)[:x]
		return hasSelfRef
	} else {
		return hasSelfRef
	}
}

// cleans up a quoted string found in card text so its easier to parse
func CleanupToken(t *model.QuotedToken) {
	*t = strings.TrimSpace(*t)
	*t = strings.ReplaceAll(*t, "\".", "")
	*t = strings.ReplaceAll(*t, "\".", "")
	*t = strings.ReplaceAll(*t, "'.", "")
	*t = strings.ReplaceAll(*t, "',", "")

	*t = strings.Trim(*t, "'")
	*t = strings.Trim(*t, "\"")
}
