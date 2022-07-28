package model

type CardSuggestions struct {
	NamedMaterials *[]Card     `json:"namedMaterials"`
	Decks          *[]DeckList `json:"decks"`
}
