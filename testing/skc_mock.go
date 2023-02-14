package testing

import (
	"log"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

type SKCDatabaseAccessObjectMock struct{}

func (mock SKCDatabaseAccessObjectMock) GetSKCDBVersion() (string, error) {
	log.Fatalln("Method not mocked")
	return "", nil
}

func (mock SKCDatabaseAccessObjectMock) FindDesiredCardInDBUsingID(cardID string) (*model.Card, *model.APIError) {
	log.Fatalln("Method not mocked")
	return nil, nil
}

func (mock SKCDatabaseAccessObjectMock) FindDesiredCardInDBUsingMultipleCardIDs(cards []string) (model.DeckListContents, model.APIError) {
	log.Fatalln("Method not mocked")
	return model.DeckListContents{}, model.APIError{}
}

func (mock SKCDatabaseAccessObjectMock) FindDesiredCardInDBUsingName(cardName string) (model.Card, error) {
	if card, isPresent := CardMocks[cardName]; isPresent {
		return card, nil
	} else {
		return model.Card{}, ErrorMock{}
	}
}

func (imp SKCDatabaseAccessObjectMock) FindOccurrenceOfCardNameInAllCardEffect(cardName string, cardId string) (*[]model.Card, *model.APIError) {
	log.Fatalln("Method not mocked")
	return nil, nil
}

type ErrorMock struct {
}

func (e ErrorMock) Error() string {
	return "mock error"
}
