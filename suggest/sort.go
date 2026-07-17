package suggest

import (
	"cmp"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func SortCardReferences(ccIDs map[string]uint32) func(a, b model.CardReference) int {
	return func(a, b model.CardReference) int {
		switch {
		case a.Occurrences != b.Occurrences:
			return cmp.Compare(b.Occurrences, a.Occurrences)
		case a.Card.GetColor() != b.Card.GetColor():
			return cmp.Compare(ccIDs[a.Card.GetColor()], ccIDs[b.Card.GetColor()])
		default:
			return cmp.Compare(a.Card.GetName(), b.Card.GetName())
		}
	}
}
