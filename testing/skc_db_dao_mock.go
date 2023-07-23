package testing

import (
	"log"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	notMocked = "Method not mocked"
)

type SKCDatabaseAccessObjectMock struct{}

func (mock SKCDatabaseAccessObjectMock) GetSKCDBVersion() (string, error) {
	log.Fatalln(notMocked)
	return "", nil
}

func (mock SKCDatabaseAccessObjectMock) GetCardColorIDs() (map[string]int, *model.APIError) {
	ids := map[string]int{
		"Normal":           1,
		"Effect":           2,
		"Fusion":           3,
		"Ritual":           4,
		"Synchro":          5,
		"Xyz":              6,
		"Pendulum-Normal":  7,
		"Pendulum-Effect":  8,
		"Pendulum-Ritual":  9,
		"Pendulum-Fusion":  10,
		"Pendulum-Xyz":     11,
		"Pendulum-Synchro": 12,
		"Link":             13,
		"Spell":            14,
		"Trap":             15,
		"Token":            16,
	}

	return ids, nil
}

func (mock SKCDatabaseAccessObjectMock) FindDesiredCardInDBUsingID(cardID string) (*model.Card, *model.APIError) {
	log.Fatalln(notMocked)
	return nil, nil
}

func (mock SKCDatabaseAccessObjectMock) FindDesiredCardInDBUsingMultipleCardIDs(cards []string) (model.CardDataMap, model.APIError) {
	log.Fatalln(notMocked)
	return model.CardDataMap{}, model.APIError{}
}

func (mock SKCDatabaseAccessObjectMock) FindDesiredCardInDBUsingName(cardName string) (model.Card, error) {
	if card, isPresent := CardMocks[cardName]; isPresent {
		return card, nil
	} else {
		return model.Card{}, ErrorMock{}
	}
}

func (imp SKCDatabaseAccessObjectMock) FindOccurrenceOfCardNameInAllCardEffect(cardName string, cardId string) ([]model.Card, *model.APIError) {
	log.Fatalln(notMocked)
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
