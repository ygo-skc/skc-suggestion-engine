package testing

import (
	"log"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

type SKCSuggestionEngineDAOImplementation struct{}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetSKCSuggestionDBVersion() (string, error) {
	return "1.0.0", nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) InsertDeckList(deckList model.DeckList) {
	log.Fatalln("InsertDeckList() not mocked")
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetDeckList(deckID string) (*model.DeckList, *model.APIError) {
	log.Fatalln("GetDeckList() not mocked")
	return nil, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetDecksThatFeatureCards(cardIDs []string) (*[]model.DeckList, *model.APIError) {
	deck := []model.DeckList{}

	return &deck, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) InsertTrafficData(ta model.TrafficAnalysis) *model.APIError {
	log.Fatalln("InsertTrafficData() not mocked")
	return nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetTrafficData(resourceName string, from time.Time, to time.Time) ([]model.Trending, *model.APIError) {
	log.Fatalln("GetTrafficData() not mocked")
	return nil, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) IsBlackListed(blackListType string, blackListPhrase string) (bool, *model.APIError) {
	log.Fatalln("IsBlackListed() not mocked")
	return false, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetCardOfTheDay(date string) (*string, *model.APIError) {
	log.Fatalln("GetCardOfTheDay() not mocked")
	return nil, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) InsertCardOfTheDay(cotd model.CardOfTheDay) *model.APIError {
	log.Fatalln("InsertCardOfTheDay() not mocked")
	return nil
}
