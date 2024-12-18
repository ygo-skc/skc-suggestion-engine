package testing

import (
	"context"
	"log"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	notMocked = "Method not mocked"
)

type SKCDatabaseAccessObjectMock struct{}

func (mock SKCDatabaseAccessObjectMock) GetSKCDBVersion(ctx context.Context) (string, error) {
	log.Fatalln(notMocked)
	return "", nil
}

func (mock SKCDatabaseAccessObjectMock) GetCardColorIDs(ctx context.Context) (map[string]int, *model.APIError) {
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

func (mock SKCDatabaseAccessObjectMock) GetDesiredCardInDBUsingID(ctx context.Context, cardID string) (model.Card, *model.APIError) {
	log.Fatalln(notMocked)
	return model.Card{}, nil
}

func (mock SKCDatabaseAccessObjectMock) GetDesiredCardInDBUsingMultipleCardIDs(ctx context.Context, cards []string) (model.BatchCardData[model.CardIDs], *model.APIError) {
	log.Fatalln(notMocked)
	return model.BatchCardData[model.CardIDs]{}, nil
}

func (mock SKCDatabaseAccessObjectMock) GetDesiredCardsFromDBUsingMultipleCardNames(ctx context.Context, cardNames []string) (model.BatchCardData[model.CardNames], *model.APIError) {
	found, notFound := make(model.CardDataMap, 0), make(model.CardNames, 0)
	for _, cardName := range cardNames {
		if card, isPresent := CardMocks[cardName]; isPresent {
			found[card.CardName] = card
		} else {
			notFound = append(notFound, cardName)
		}
	}

	return model.BatchCardData[model.CardNames]{CardInfo: found, UnknownResources: notFound}, nil
}
func (imp SKCDatabaseAccessObjectMock) GetCardsFoundInProduct(ctx context.Context, productId string) (model.BatchCardData[model.CardIDs], *model.APIError) {
	log.Fatalln("GetInArchetypeSupportUsingCardName() not mocked")
	return model.BatchCardData[model.CardIDs]{}, nil
}

func (imp SKCDatabaseAccessObjectMock) GetOccurrenceOfCardNameInAllCardEffect(ctx context.Context, cardName string, cardId string) ([]model.Card, *model.APIError) {
	log.Fatalln(notMocked)
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetInArchetypeSupportUsingCardName(ctx context.Context, archetypeName string) ([]model.Card, *model.APIError) {
	log.Fatalln("GetInArchetypeSupportUsingCardName() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetInArchetypeSupportUsingCardText(ctx context.Context, archetypeName string) ([]model.Card, *model.APIError) {
	log.Fatalln("GetInArchetypeSupportUsingCardText() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetArchetypeExclusionsUsingCardText(ctx context.Context, archetypeName string) ([]model.Card, *model.APIError) {
	log.Fatalln("GetArchetypeExclusionsUsingCardText() not mocked")
	return nil, nil
}
func (imp SKCDatabaseAccessObjectMock) GetDesiredProductInDBUsingID(ctx context.Context, productID string) (*model.Product, *model.APIError) {
	log.Fatalln("GetDesiredProductInDBUsingID() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetDesiredProductInDBUsingMultipleProductIDs(ctx context.Context, cards []string) (model.BatchProductData[model.ProductIDs], *model.APIError) {
	log.Fatalln("GetDesiredProductInDBUsingMultipleProductIDs() not mocked")
	return model.BatchProductData[model.ProductIDs]{}, nil
}

func (imp SKCDatabaseAccessObjectMock) GetRandomCard(ctx context.Context, blacklistedCards []string) (string, *model.APIError) {
	log.Fatalln("GetRandomCard() not mocked")
	return "", nil
}

type ErrorMock struct {
}

func (e ErrorMock) Error() string {
	return "mock error"
}
