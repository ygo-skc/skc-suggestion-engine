package testing

import (
	"context"

	"github.com/ygo-skc/skc-go/common/model"
	"github.com/ygo-skc/skc-go/common/ygo"
)

const (
	ni = "Not implemented"
)

type YGOCardClientMock struct{}

func (svc YGOCardClientMock) GetCardColorsProto(ctx context.Context) (*ygo.CardColors, *model.APIError) {
	ids := map[string]uint32{
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

	return &ygo.CardColors{Values: ids}, nil
}

func (svc YGOCardClientMock) GetCardByIDProto(ctx context.Context, cardID string) (*ygo.Card, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetCardByID(ctx context.Context, cardID string) (*model.YGOCard, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetCardsByIDProto(ctx context.Context, cardIDs model.CardIDs) (*ygo.Cards, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetCardsByID(ctx context.Context, cardIDs model.CardIDs) (*model.BatchCardData[model.CardIDs], *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetCardsByNameProto(ctx context.Context, cardNames model.CardNames) (*ygo.Cards, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetCardsByName(ctx context.Context, cardNames model.CardNames) (*model.BatchCardData[model.CardNames], *model.APIError) {
	found, notFound := make(model.CardDataMap, 0), make(model.CardNames, 0)
	for _, cardName := range cardNames {
		if card, isPresent := CardMocks[cardName]; isPresent {
			found[card.Name] = card
		} else {
			notFound = append(notFound, cardName)
		}
	}

	return &model.BatchCardData[model.CardNames]{CardInfo: found, UnknownResources: notFound}, nil
}

func (svc YGOCardClientMock) SearchForCardRefUsingEffectProto(ctx context.Context, cardName string, cardID string) (*ygo.CardList, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) SearchForCardRefUsingEffect(ctx context.Context, cardName string, cardID string) ([]model.YGOCard, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetArchetypalCardsUsingCardNameProto(ctx context.Context, archetype string) (*ygo.CardList, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetArchetypalCardsUsingCardName(ctx context.Context, archetype string) ([]model.YGOCard, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetExplicitArchetypalInclusionsProto(ctx context.Context, archetype string) (*ygo.CardList, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetExplicitArchetypalInclusions(ctx context.Context, archetype string) ([]model.YGOCard, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetExplicitArchetypalExclusionsProto(ctx context.Context, archetype string) (*ygo.CardList, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetExplicitArchetypalExclusions(ctx context.Context, archetype string) ([]model.YGOCard, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetRandomCardProto(ctx context.Context, blackListedIDs []string) (*ygo.Card, *model.APIError) {
	panic(ni)
}

func (svc YGOCardClientMock) GetRandomCard(ctx context.Context, blackListedIDs []string) (*model.YGOCard, *model.APIError) {
	panic(ni)
}
