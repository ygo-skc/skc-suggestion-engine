package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	skc_testing "github.com/ygo-skc/skc-suggestion-engine/testing"
)

func validateMaterialReferences(card model.Card, expectedNamedMaterials []model.CardReference, expectedMaterialArchetypes []string, assert *assert.Assertions) {
	materialString := card.GetPotentialMaterialsAsString()
	refs, archetypes := getReferences(materialString)

	assert.Len(expectedNamedMaterials, len(*refs), "Len of NamedMaterials mismatch")
	assert.Len(expectedMaterialArchetypes, len(*archetypes), "Len of MaterialArchetypes mismatch")

	assert.Equal(expectedNamedMaterials, *refs, "Expected contents of NamedMaterials slice is different than what is actually received")
	assert.Equal(expectedMaterialArchetypes, *archetypes, "Expected contents of MaterialArchetypes slice is different than what is actually received")
}

func validateReferences(card model.Card, expectedNamedReferences []model.CardReference, expectedReferencedArchetypes []string, assert *assert.Assertions) {
	materialString := card.GetPotentialMaterialsAsString()
	effectWithoutMaterial := strings.ReplaceAll(card.CardEffect, materialString, "")
	refs, archetypes := getReferences(effectWithoutMaterial)

	assert.Len(expectedNamedReferences, len(*refs), "Len of NamedReferences mismatch")
	assert.Len(expectedReferencedArchetypes, len(*archetypes), "Len of ReferencedArchetypes mismatch")

	assert.Equal(expectedNamedReferences, *refs, "Expected contents of NamedReferences slice is different than what is actually received")
	assert.Equal(expectedReferencedArchetypes, *archetypes, "Expected contents of ReferencedArchetypes slice is different than what is actually received")
}

func TestGetSuggestions(t *testing.T) {
	// setup
	assert := assert.New(t)
	skcDBInterface = skc_testing.SKCDatabaseAccessObjectMock{}
	skcSuggestionEngineDBInterface = skc_testing.SKCSuggestionEngineDAOImplementation{}

	for cardName := range skc_testing.CardSuggestionsWithSelfReferenceMock {
		mock := skc_testing.CardMocks[cardName]
		suggestions := getSuggestions(&mock)

		assert.Equal(skc_testing.CardSuggestionsWithSelfReferenceMock[cardName].NamedMaterials, suggestions.NamedMaterials, "Named Material values did not match")
		assert.Equal(skc_testing.CardSuggestionsWithSelfReferenceMock[cardName].MaterialArchetypes, suggestions.MaterialArchetypes, "Material Archetype values did not match")

		assert.Equal(skc_testing.CardSuggestionsWithoutSelfReferenceMock[cardName].NamedReferences, suggestions.NamedReferences, "Named References values did not match")
		assert.Equal(skc_testing.CardSuggestionsWithoutSelfReferenceMock[cardName].ReferencedArchetypes, suggestions.ReferencedArchetypes, "Referenced Archetype values did not match")
	}
}

func TestGetReferences(t *testing.T) {
	// setup
	assert := assert.New(t)
	skcDBInterface = skc_testing.SKCDatabaseAccessObjectMock{}

	for cardName := range skc_testing.CardSuggestionsWithSelfReferenceMock {
		validateMaterialReferences(
			skc_testing.CardMocks[cardName],
			*skc_testing.CardSuggestionsWithSelfReferenceMock[cardName].NamedMaterials,
			*skc_testing.CardSuggestionsWithSelfReferenceMock[cardName].MaterialArchetypes,
			assert,
		)

		validateReferences(
			skc_testing.CardMocks[cardName],
			*skc_testing.CardSuggestionsWithSelfReferenceMock[cardName].NamedReferences,
			*skc_testing.CardSuggestionsWithSelfReferenceMock[cardName].ReferencedArchetypes,
			assert,
		)
	}
}

func TestCleanupReference(t *testing.T) {
	assert := assert.New(t)

	baseCases := []string{"'Sunrise", "'Sunrise'", "Sunrise'"}
	for _, value := range baseCases {
		model.CleanupToken(&value)
		assert.Equal("Sunrise", value, "Expected token - after cleanup - does not equal actual value")
	}

	specialCases := []string{"Iron Core of Koa'ki Meiru", "'Iron Core of Koa'ki Meiru", "'Iron Core of Koa'ki Meiru'", "Iron Core of Koa'ki Meiru\""}
	for _, value := range specialCases {
		model.CleanupToken(&value)
		assert.Equal("Iron Core of Koa'ki Meiru", value, "Expected token - after cleanup - does not equal actual value")
	}
}