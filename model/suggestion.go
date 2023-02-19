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
	Card         *Card  `json:"card"`
	ReferencedBy []Card `json:"referencedBy"`
	MaterialFor  []Card `json:"materialFor"`
}
