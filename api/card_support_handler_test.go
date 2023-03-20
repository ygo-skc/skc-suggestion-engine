package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	skc_testing "github.com/ygo-skc/skc-suggestion-engine/testing"
)

var (
	// this object is mocking what would return from the DB prior to organizing the references by material or generic non material
	cardReferenceSubjects = map[string][]model.Card{
		"Dark Magician":                   {skc_testing.CardMocks["Magicians' Souls"], skc_testing.CardMocks["Dark Paladin"]},
		"Hamon, Lord of Striking Thunder": {skc_testing.CardMocks["Armityle the Chaos Phantasm"], skc_testing.CardMocks["Armityle the Chaos Phantasm - Phantom of Fury"]},
		"Elemental HERO Neos":             {skc_testing.CardMocks["Neos Wiseman"], skc_testing.CardMocks["Elemental HERO Air Neos"]},
	}
)

func TestDetermineSupportCards(t *testing.T) {
	// setup
	assert := assert.New(t)

	for cardName, references := range cardReferenceSubjects {
		cardMock := skc_testing.CardMocks[cardName]
		assert.Equal(cardName, cardMock.CardName, fmt.Sprintf("Mock not setup for %s", cardName))

		actualReferencedBy, actualMaterialFor := determineSupportCards(cardMock, references)

		expectedReferencedBy := skc_testing.ExpectedSupportCardsMocks[cardName].ReferencedBy
		expectedMaterialFor := skc_testing.ExpectedSupportCardsMocks[cardName].MaterialFor

		assert.Len(actualReferencedBy, len(expectedReferencedBy), "Len of ReferencedBy array is incorrect")
		assert.Len(actualMaterialFor, len(expectedMaterialFor), "Len of MaterialFor array is incorrect")

		assert.Equal(actualReferencedBy, expectedReferencedBy, "Expected contents of ReferencedBy slice is different than what is actually received")
		assert.Equal(actualMaterialFor, expectedMaterialFor, "Expected contents of MaterialFor slice is different than what is actually received")
	}
}
