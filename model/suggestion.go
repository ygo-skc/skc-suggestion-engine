package model

type CardReference struct {
	Occurrences int  `json:"occurrences"`
	Card        Card `json:"card"`
}

type CardSuggestions struct {
	NamedMaterials  *[]CardReference `json:"namedMaterials"`
	NamedReferences *[]CardReference `json:"namedReferences"`
	Decks           *[]DeckList      `json:"decks"`
}
