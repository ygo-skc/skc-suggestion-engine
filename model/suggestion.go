package model

type CardReference struct {
	Occurrences int  `json:"occurrences"`
	Card        Card `json:"card"`
}

type CardSuggestions struct {
	Card                 *Card            `json:"card"`
	HasSelfReference     bool             `json:"hasSelfReference"`
	NamedMaterials       *[]CardReference `json:"namedMaterials"`
	NamedReferences      *[]CardReference `json:"namedReferences"`
	MaterialArchetypes   *[]string        `json:"materialArchetypes"`
	ReferencedArchetypes *[]string        `json:"referencedArchetypes"`
	Decks                *[]DeckList      `json:"decks"`
}

type CardSupport struct {
	Card    *Card            `json:"card"`
	Support *[]CardReference `json:"support"`
}

// looks for a self reference, if a self reference is found it is removed from original slice
// this method returns true if a self reference is found
func RemoveSelfReference(self string, cr *[]CardReference) bool {
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
