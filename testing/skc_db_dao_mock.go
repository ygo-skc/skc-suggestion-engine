package testing

import (
	"context"
	"log"

	cModel "github.com/ygo-skc/skc-go/common/model"
)

const (
	notMocked = "Method not mocked"
)

type SKCDatabaseAccessObjectMock struct{}

func (mock SKCDatabaseAccessObjectMock) GetSKCDBVersion(ctx context.Context) (string, error) {
	log.Fatalln(notMocked)
	return "", nil
}

func (mock SKCDatabaseAccessObjectMock) GetCardColorIDs(ctx context.Context) (map[string]int, *cModel.APIError) {
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

func (mock SKCDatabaseAccessObjectMock) GetDesiredCardInDBUsingID(ctx context.Context, cardID string) (cModel.Card, *cModel.APIError) {
	log.Fatalln(notMocked)
	return cModel.Card{}, nil
}

func (mock SKCDatabaseAccessObjectMock) GetDesiredCardInDBUsingMultipleCardIDs(ctx context.Context, cards []string) (cModel.BatchCardData[cModel.CardIDs], *cModel.APIError) {
	log.Fatalln(notMocked)
	return cModel.BatchCardData[cModel.CardIDs]{}, nil
}

func (mock SKCDatabaseAccessObjectMock) GetDesiredCardsFromDBUsingMultipleCardNames(ctx context.Context, cardNames []string) (cModel.BatchCardData[cModel.CardNames], *cModel.APIError) {
	found, notFound := make(cModel.CardDataMap, 0), make(cModel.CardNames, 0)
	for _, cardName := range cardNames {
		if card, isPresent := CardMocks[cardName]; isPresent {
			found[card.CardName] = card
		} else {
			notFound = append(notFound, cardName)
		}
	}

	return cModel.BatchCardData[cModel.CardNames]{CardInfo: found, UnknownResources: notFound}, nil
}
func (imp SKCDatabaseAccessObjectMock) GetCardsFoundInProduct(ctx context.Context, productId string) (cModel.BatchCardData[cModel.CardIDs], *cModel.APIError) {
	log.Fatalln("GetInArchetypeSupportUsingCardName() not mocked")
	return cModel.BatchCardData[cModel.CardIDs]{}, nil
}

func (imp SKCDatabaseAccessObjectMock) GetOccurrenceOfCardNameInAllCardEffect(ctx context.Context, cardName string, cardId string) ([]cModel.Card, *cModel.APIError) {
	log.Fatalln(notMocked)
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetInArchetypeSupportUsingCardName(ctx context.Context, archetypeName string) ([]cModel.Card, *cModel.APIError) {
	log.Fatalln("GetInArchetypeSupportUsingCardName() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetInArchetypeSupportUsingCardText(ctx context.Context, archetypeName string) ([]cModel.Card, *cModel.APIError) {
	log.Fatalln("GetInArchetypeSupportUsingCardText() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetArchetypeExclusionsUsingCardText(ctx context.Context, archetypeName string) ([]cModel.Card, *cModel.APIError) {
	log.Fatalln("GetArchetypeExclusionsUsingCardText() not mocked")
	return nil, nil
}
func (imp SKCDatabaseAccessObjectMock) GetDesiredProductInDBUsingID(ctx context.Context, productID string) (*cModel.Product, *cModel.APIError) {
	log.Fatalln("GetDesiredProductInDBUsingID() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetDesiredProductInDBUsingMultipleProductIDs(ctx context.Context, cards []string) (cModel.BatchProductData[cModel.ProductIDs], *cModel.APIError) {
	log.Fatalln("GetDesiredProductInDBUsingMultipleProductIDs() not mocked")
	return cModel.BatchProductData[cModel.ProductIDs]{}, nil
}

func (imp SKCDatabaseAccessObjectMock) GetRandomCard(ctx context.Context, blacklistedCards []string) (string, *cModel.APIError) {
	log.Fatalln("GetRandomCard() not mocked")
	return "", nil
}

type ErrorMock struct {
}

func (e ErrorMock) Error() string {
	return "mock error"
}
