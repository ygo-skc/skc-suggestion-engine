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

func (mock SKCDatabaseAccessObjectMock) GetDesiredCardsFromDBUsingMultipleCardNames(ctx context.Context, cardNames []string) (cModel.BatchCardData[cModel.CardNames], *cModel.APIError) {
	found, notFound := make(cModel.CardDataMap, 0), make(cModel.CardNames, 0)
	for _, cardName := range cardNames {
		if card, isPresent := CardMocks[cardName]; isPresent {
			found[card.Name] = card
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

func (imp SKCDatabaseAccessObjectMock) GetOccurrenceOfCardNameInAllCardEffect(ctx context.Context, cardName string, cardId string) ([]cModel.YGOCard, *cModel.APIError) {
	log.Fatalln(notMocked)
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetInArchetypeSupportUsingCardName(ctx context.Context, archetypeName string) ([]cModel.YGOCard, *cModel.APIError) {
	log.Fatalln("GetInArchetypeSupportUsingCardName() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetInArchetypeSupportUsingCardText(ctx context.Context, archetypeName string) ([]cModel.YGOCard, *cModel.APIError) {
	log.Fatalln("GetInArchetypeSupportUsingCardText() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetArchetypeExclusionsUsingCardText(ctx context.Context, archetypeName string) ([]cModel.YGOCard, *cModel.APIError) {
	log.Fatalln("GetArchetypeExclusionsUsingCardText() not mocked")
	return nil, nil
}
func (imp SKCDatabaseAccessObjectMock) GetDesiredProductInDBUsingID(ctx context.Context, productID string) (*cModel.YGOProduct, *cModel.APIError) {
	log.Fatalln("GetDesiredProductInDBUsingID() not mocked")
	return nil, nil
}

func (imp SKCDatabaseAccessObjectMock) GetDesiredProductInDBUsingMultipleProductIDs(ctx context.Context, cards []string) (cModel.BatchProductData[cModel.ProductIDs], *cModel.APIError) {
	log.Fatalln("GetDesiredProductInDBUsingMultipleProductIDs() not mocked")
	return cModel.BatchProductData[cModel.ProductIDs]{}, nil
}

type ErrorMock struct {
}

func (e ErrorMock) Error() string {
	return "mock error"
}
