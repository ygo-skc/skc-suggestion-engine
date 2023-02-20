package testing

import "github.com/ygo-skc/skc-suggestion-engine/model"

type SKCSuggestionEngineDAOImplementation struct{}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetSKCSuggestionDBVersion() (string, error) {
	return "1.0.0", nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetDecksThatFeatureCards(cardIDs []string) (*[]model.DeckList, *model.APIError) {
	deck := []model.DeckList{}

	return &deck, nil
}
