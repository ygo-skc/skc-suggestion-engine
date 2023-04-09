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

func (imp SKCDatabaseAccessObjectMock) FindOccurrenceOfCardNameInAllCardEffect(cardName string, cardId string) ([]model.Card, *model.APIError) {
	log.Fatalln("Method not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) FindInArchetypeSupportUsingCardName(archetypeName string) ([]model.Card, *model.APIError) {
	log.Fatalln("FindInArchetypeSupportUsingCardName() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) FindInArchetypeSupportUsingCardText(archetypeName string) ([]model.Card, *model.APIError) {
	log.Fatalln("FindInArchetypeSupportUsingCardText() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) FindArchetypeExclusionsUsingCardText(archetypeName string) ([]model.Card, *model.APIError) {
	log.Fatalln("FindArchetypeExclusionsUsingCardText() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetRandomCard() (string, *model.APIError) {
	log.Fatalln("FindArchetypeExclusionsUsingCardText() not mocked")
	return "", nil
}

type ErrorMock struct {
}

func (e ErrorMock) Error() string {
	return "mock error"
}
