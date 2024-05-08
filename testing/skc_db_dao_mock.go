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

func (mock SKCDatabaseAccessObjectMock) GetDesiredCardInDBUsingID(cardID string) (model.Card, *model.APIError) {
	log.Fatalln(notMocked)
	return model.Card{}, nil
}

func (mock SKCDatabaseAccessObjectMock) GetDesiredCardInDBUsingMultipleCardIDs(cards []string) (*model.BatchCardData[model.CardIDs], *model.APIError) {
	log.Fatalln(notMocked)
	return &model.BatchCardData[model.CardIDs]{}, nil
}

func (mock SKCDatabaseAccessObjectMock) GetDesiredCardsFromDBUsingMultipleCardNames(cardNames []string) (*model.BatchCardData[model.CardNames], *model.APIError) {
	found, notFound := make(model.CardDataMap, 0), make(model.CardNames, 0)
	for _, cardName := range cardNames {
		if card, isPresent := CardMocks[cardName]; isPresent {
			found[card.CardName] = card
		} else {
			notFound = append(notFound, cardName)
		}
	}

	return &model.BatchCardData[model.CardNames]{CardInfo: found, UnknownResources: notFound}, nil
}

func (imp SKCDatabaseAccessObjectMock) GetOccurrenceOfCardNameInAllCardEffect(cardName string, cardId string) ([]model.Card, *model.APIError) {
	log.Fatalln(notMocked)
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetInArchetypeSupportUsingCardName(archetypeName string) ([]model.Card, *model.APIError) {
	log.Fatalln("GetInArchetypeSupportUsingCardName() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetInArchetypeSupportUsingCardText(archetypeName string) ([]model.Card, *model.APIError) {
	log.Fatalln("GetInArchetypeSupportUsingCardText() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetArchetypeExclusionsUsingCardText(archetypeName string) ([]model.Card, *model.APIError) {
	log.Fatalln("GetArchetypeExclusionsUsingCardText() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetDesiredProductInDBUsingMultipleProductIDs(cards []string) (*model.BatchProductData[model.ProductIDs], *model.APIError) {
	log.Fatalln("GetDesiredProductInDBUsingMultipleProductIDs() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetRandomCard() (string, *model.APIError) {
	log.Fatalln("GetArchetypeExclusionsUsingCardText() not mocked")
	return "", nil
}

type ErrorMock struct {
}

func (e ErrorMock) Error() string {
	return "mock error"
}
