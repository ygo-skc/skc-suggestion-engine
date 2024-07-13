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
		"Dark Magician":                   {skc_testing.CardMocks["Magicians' Souls"], skc_testing.CardMocks["Dark Paladin"], skc_testing.CardMocks["The Dark Magicians"]},
		"Hamon, Lord of Striking Thunder": {skc_testing.CardMocks["Armityle the Chaos Phantasm"], skc_testing.CardMocks["Armityle the Chaos Phantasm - Phantom of Fury"]},
		"Elemental HERO Neos":             {skc_testing.CardMocks["Neos Wiseman"], skc_testing.CardMocks["Elemental HERO Air Neos"]},
	}

	// expected output when fetching support cards
	expectedSupportCardsMocks = map[string]model.CardSupport{
		"Dark Magician": {
			ReferencedBy: []model.CardReference{
				{Card: skc_testing.CardMocks["Magicians' Souls"], Occurrences: 1},
				{Card: skc_testing.CardMocks["The Dark Magicians"], Occurrences: 1}},
			MaterialFor: []model.CardReference{
				{Card: skc_testing.CardMocks["Dark Paladin"], Occurrences: 1},
				{Card: skc_testing.CardMocks["The Dark Magicians"], Occurrences: 1}},
		},
		"Hamon, Lord of Striking Thunder": {
			ReferencedBy: []model.CardReference{},
			MaterialFor: []model.CardReference{
				{Card: skc_testing.CardMocks["Armityle the Chaos Phantasm"], Occurrences: 1},
				{Card: skc_testing.CardMocks["Armityle the Chaos Phantasm - Phantom of Fury"], Occurrences: 1}},
		},
		"Elemental HERO Neos": {
			ReferencedBy: []model.CardReference{{Card: skc_testing.CardMocks["Neos Wiseman"], Occurrences: 1}},
			MaterialFor:  []model.CardReference{{Card: skc_testing.CardMocks["Elemental HERO Air Neos"], Occurrences: 1}},
		},
	}
)

func TestDetermineSupportCards(t *testing.T) {
	// setup
	assert := assert.New(t)

	for cardName, references := range cardReferenceSubjects {
		cardMock := skc_testing.CardMocks[cardName]
		assert.Equal(cardName, cardMock.CardName, fmt.Sprintf("Mock not setup for %s", cardName))

		actualReferencedBy, actualMaterialFor := determineSupportCards(cardMock, references)

		expectedReferencedBy := expectedSupportCardsMocks[cardName].ReferencedBy
		expectedMaterialFor := expectedSupportCardsMocks[cardName].MaterialFor

		assert.Len(actualReferencedBy, len(expectedReferencedBy), "Len of ReferencedBy array is incorrect")
		assert.Len(actualMaterialFor, len(expectedMaterialFor), "Len of MaterialFor array is incorrect")

		assert.Equal(actualReferencedBy, expectedReferencedBy, "Expected contents of ReferencedBy slice is different than what is actually received")
		assert.Equal(actualMaterialFor, expectedMaterialFor, "Expected contents of MaterialFor slice is different than what is actually received")
	}
}
