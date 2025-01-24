package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	cModel "github.com/ygo-skc/skc-go/common/model"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	skc_testing "github.com/ygo-skc/skc-suggestion-engine/testing"
)

func validateMaterialReferences(card cModel.Card, expectedNamedMaterials []model.CardReference, expectedMaterialArchetypes []string, assert *assert.Assertions) {
	materialString := card.GetPotentialMaterialsAsString()
	refs, archetypes := getReferences(skc_testing.TestContext, materialString)

	assert.Len(expectedNamedMaterials, len(refs), "Len of NamedMaterials mismatch")
	assert.Len(expectedMaterialArchetypes, len(archetypes), "Len of MaterialArchetypes mismatch")

	assert.Equal(expectedNamedMaterials, refs, "Expected contents of NamedMaterials slice is different than what is actually received")
	assert.Equal(expectedMaterialArchetypes, archetypes, "Expected contents of MaterialArchetypes slice is different than what is actually received")
}

func validateReferences(card cModel.Card, expectedNamedReferences []model.CardReference, expectedReferencedArchetypes []string, assert *assert.Assertions) {
	materialString := card.GetPotentialMaterialsAsString()
	effectWithoutMaterial := strings.ReplaceAll(card.CardEffect, materialString, "")
	refs, archetypes := getReferences(skc_testing.TestContext, effectWithoutMaterial)

	assert.Len(expectedNamedReferences, len(refs), "Len of NamedReferences mismatch")
	assert.Len(expectedReferencedArchetypes, len(archetypes), "Len of ReferencedArchetypes mismatch")

	assert.Equal(expectedNamedReferences, refs, "Expected contents of NamedReferences slice is different than what is actually received")
	assert.Equal(expectedReferencedArchetypes, archetypes, "Expected contents of ReferencedArchetypes slice is different than what is actually received")
}

func TestGetSuggestions(t *testing.T) {
	// setup
	assert := assert.New(t)
	skcDBInterface = skc_testing.SKCDatabaseAccessObjectMock{}
	skcSuggestionEngineDBInterface = skc_testing.SKCSuggestionEngineDAOImplementation{}

	ccIDs, _ := skcDBInterface.GetCardColorIDs(skc_testing.TestContext)
	for cardName := range cardSuggestionsWithSelfReferenceMock {
		mock := skc_testing.CardMocks[cardName]
		suggestions := getCardSuggestions(skc_testing.TestContext, mock, ccIDs)

		assert.Equal(cardSuggestionsWithSelfReferenceMock[cardName].NamedMaterials, suggestions.NamedMaterials, "Named Material values did not match")
		assert.Equal(cardSuggestionsWithSelfReferenceMock[cardName].MaterialArchetypes, suggestions.MaterialArchetypes, "Material Archetype values did not match")

		assert.Equal(cardSuggestionsWithoutSelfReferenceMock[cardName].NamedReferences, suggestions.NamedReferences, "Named References values did not match")
		assert.Equal(cardSuggestionsWithoutSelfReferenceMock[cardName].ReferencedArchetypes, suggestions.ReferencedArchetypes, "Referenced Archetype values did not match")
	}
}

func TestCleanupReference(t *testing.T) {
	assert := assert.New(t)

	baseCases := []string{"'Sunrise", "'Sunrise'", "Sunrise'"}
	for _, value := range baseCases {
		cModel.CleanupToken(&value)
		assert.Equal("Sunrise", value, "Expected token - after cleanup - does not equal actual value")
	}

	specialCases := []string{"Iron Core of Koa'ki Meiru", "'Iron Core of Koa'ki Meiru", "'Iron Core of Koa'ki Meiru'", "Iron Core of Koa'ki Meiru\""}
	for _, value := range specialCases {
		cModel.CleanupToken(&value)
		assert.Equal("Iron Core of Koa'ki Meiru", value, "Expected token - after cleanup - does not equal actual value")
	}
}

var (
	cardSuggestionsWithSelfReferenceMock = map[string]model.CardSuggestions{
		"Elemental HERO Sunrise": {
			NamedMaterials:       []model.CardReference{},
			MaterialArchetypes:   []string{"HERO"},
			NamedReferences:      []model.CardReference{{Occurrences: 1, Card: skc_testing.CardMocks["Elemental HERO Sunrise"]}, {Occurrences: 1, Card: skc_testing.CardMocks["Miracle Fusion"]}},
			ReferencedArchetypes: []string{"HERO"},
		},
		"Gem-Knight Master Diamond": {
			NamedMaterials:       []model.CardReference{},
			MaterialArchetypes:   []string{"Gem-Knight"},
			NamedReferences:      []model.CardReference{},
			ReferencedArchetypes: []string{"Gem-", "Gem-Knight"},
		},
		"A-to-Z-Dragon Buster Cannon": {
			NamedMaterials:       []model.CardReference{{Occurrences: 1, Card: skc_testing.CardMocks["ABC-Dragon Buster"]}, {Occurrences: 1, Card: skc_testing.CardMocks["XYZ-Dragon Cannon"]}},
			MaterialArchetypes:   []string{},
			NamedReferences:      []model.CardReference{{Occurrences: 1, Card: skc_testing.CardMocks["ABC-Dragon Buster"]}, {Occurrences: 1, Card: skc_testing.CardMocks["XYZ-Dragon Cannon"]}, {Occurrences: 1, Card: skc_testing.CardMocks["Polymerization"]}},
			ReferencedArchetypes: []string{},
		},
		"The Legendary Fisherman II": {
			NamedMaterials:       []model.CardReference{},
			MaterialArchetypes:   []string{},
			NamedReferences:      []model.CardReference{{Occurrences: 1, Card: skc_testing.CardMocks["The Legendary Fisherman"]}, {Occurrences: 1, Card: skc_testing.CardMocks["Umi"]}},
			ReferencedArchetypes: []string{},
		},
		"Armityle the Chaos Phantasm": {
			NamedMaterials: []model.CardReference{
				{Occurrences: 1, Card: skc_testing.CardMocks["Hamon, Lord of Striking Thunder"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["Raviel, Lord of Phantasms"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["Uria, Lord of Searing Flames"]},
			},
			MaterialArchetypes: []string{},
			NamedReferences: []model.CardReference{
				{Occurrences: 1, Card: skc_testing.CardMocks["Polymerization"]},
			},
			ReferencedArchetypes: []string{},
		},
		"Armityle the Chaos Phantasm - Phantom of Fury": {
			NamedMaterials: []model.CardReference{
				{Occurrences: 1, Card: skc_testing.CardMocks["Hamon, Lord of Striking Thunder"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["Raviel, Lord of Phantasms"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["Uria, Lord of Searing Flames"]},
			},
			MaterialArchetypes: []string{},
			NamedReferences: []model.CardReference{
				{Occurrences: 2, Card: skc_testing.CardMocks["Armityle the Chaos Phantasm"]},
			},
			ReferencedArchetypes: []string{},
		},
		"King Dragun": {
			NamedMaterials: []model.CardReference{
				{Occurrences: 1, Card: skc_testing.CardMocks["Divine Dragon Ragnarok"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["Lord of D."]},
			},
			MaterialArchetypes:   []string{},
			NamedReferences:      []model.CardReference{},
			ReferencedArchetypes: []string{},
		},
		"Great Mammoth of Goldfine": {
			NamedMaterials: []model.CardReference{
				{Occurrences: 1, Card: skc_testing.CardMocks["Dragon Zombie"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["The Snake Hair"]},
			},
			MaterialArchetypes:   []string{},
			NamedReferences:      []model.CardReference{},
			ReferencedArchetypes: []string{},
		},
		"Elemental HERO Stratos": {
			NamedMaterials:     []model.CardReference{},
			MaterialArchetypes: []string{},
			NamedReferences:    []model.CardReference{},
			ReferencedArchetypes: []string{
				"HERO",
			},
		},
	}

	cardSuggestionsWithoutSelfReferenceMock = map[string]model.CardSuggestions{
		"Elemental HERO Sunrise": {
			NamedMaterials:       []model.CardReference{},
			MaterialArchetypes:   []string{"HERO"},
			NamedReferences:      []model.CardReference{{Occurrences: 1, Card: skc_testing.CardMocks["Miracle Fusion"]}},
			ReferencedArchetypes: []string{"HERO"},
		},
		"Gem-Knight Master Diamond": {
			NamedMaterials:       []model.CardReference{},
			MaterialArchetypes:   []string{"Gem-Knight"},
			NamedReferences:      []model.CardReference{},
			ReferencedArchetypes: []string{"Gem-", "Gem-Knight"},
		},
		"A-to-Z-Dragon Buster Cannon": {
			NamedMaterials:       []model.CardReference{{Occurrences: 1, Card: skc_testing.CardMocks["ABC-Dragon Buster"]}, {Occurrences: 1, Card: skc_testing.CardMocks["XYZ-Dragon Cannon"]}},
			MaterialArchetypes:   []string{},
			NamedReferences:      []model.CardReference{{Occurrences: 1, Card: skc_testing.CardMocks["ABC-Dragon Buster"]}, {Occurrences: 1, Card: skc_testing.CardMocks["XYZ-Dragon Cannon"]}, {Occurrences: 1, Card: skc_testing.CardMocks["Polymerization"]}},
			ReferencedArchetypes: []string{},
		},
		"The Legendary Fisherman II": {
			NamedMaterials:       []model.CardReference{},
			MaterialArchetypes:   []string{},
			NamedReferences:      []model.CardReference{{Occurrences: 1, Card: skc_testing.CardMocks["The Legendary Fisherman"]}, {Occurrences: 1, Card: skc_testing.CardMocks["Umi"]}},
			ReferencedArchetypes: []string{},
		},
		"Armityle the Chaos Phantasm": {
			NamedMaterials: []model.CardReference{
				{Occurrences: 1, Card: skc_testing.CardMocks["Hamon, Lord of Striking Thunder"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["Raviel, Lord of Phantasms"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["Uria, Lord of Searing Flames"]},
			},
			MaterialArchetypes: []string{},
			NamedReferences: []model.CardReference{
				{Occurrences: 1, Card: skc_testing.CardMocks["Polymerization"]},
			},
			ReferencedArchetypes: []string{},
		},
		"Armityle the Chaos Phantasm - Phantom of Fury": {
			NamedMaterials: []model.CardReference{
				{Occurrences: 1, Card: skc_testing.CardMocks["Hamon, Lord of Striking Thunder"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["Raviel, Lord of Phantasms"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["Uria, Lord of Searing Flames"]},
			},
			MaterialArchetypes: []string{},
			NamedReferences: []model.CardReference{
				{Occurrences: 2, Card: skc_testing.CardMocks["Armityle the Chaos Phantasm"]},
			},
			ReferencedArchetypes: []string{},
		},
		"King Dragun": {
			NamedMaterials: []model.CardReference{
				{Occurrences: 1, Card: skc_testing.CardMocks["Divine Dragon Ragnarok"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["Lord of D."]},
			},
			MaterialArchetypes:   []string{},
			NamedReferences:      []model.CardReference{},
			ReferencedArchetypes: []string{},
		},
		"Great Mammoth of Goldfine": {
			NamedMaterials: []model.CardReference{
				{Occurrences: 1, Card: skc_testing.CardMocks["Dragon Zombie"]},
				{Occurrences: 1, Card: skc_testing.CardMocks["The Snake Hair"]},
			},
			MaterialArchetypes:   []string{},
			NamedReferences:      []model.CardReference{},
			ReferencedArchetypes: []string{},
		},
		"Elemental HERO Stratos": {
			NamedMaterials:     []model.CardReference{},
			MaterialArchetypes: []string{},
			NamedReferences:    []model.CardReference{},
			ReferencedArchetypes: []string{
				"HERO",
			},
		},
	}
)
