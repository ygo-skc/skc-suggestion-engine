package model

type CardSuggestions struct {
	NamedMaterials  *[]Card     `json:"namedMaterials"`
	NamedReferences *[]Card     `json:"namedReferences"`
	Decks           *[]DeckList `json:"decks"`
}
