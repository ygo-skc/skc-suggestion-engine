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

	assert.Len(*refs, 0, "Len of material refs mismatch")
	assert.Len(*archetypes, 1, "Len of material archetypes mismatch")

	assert.Equal(*refs, expectedNamedMaterials)
	assert.Equal(*archetypes, expectedMaterialArchetypes, "Expected size of archetype slice is different than what is actually received")
}

func validateReferences(card model.Card, expectedNamedReferences []model.CardReference, expectedReferencedArchetypes []string, assert *assert.Assertions) {
	materialString := card.GetPotentialMaterialsAsString()
	effectWithoutMaterial := strings.ReplaceAll(card.CardEffect, materialString, "")
	refs, archetypes := getReferences(effectWithoutMaterial)

	assert.Len(*refs, 2, "Len of refs mismatch")
	assert.Len(*archetypes, 1, "Len of archetypes mismatch")

	assert.Equal(*refs, expectedNamedReferences)
	assert.Equal(*archetypes, expectedReferencedArchetypes, "Expected size of archetype slice is different than what is actually received")
}

func TestGetReferences(t *testing.T) {
	assert := assert.New(t)
	skcDBInterface = skc_testing.SKCDatabaseAccessObjectMock{}

	expectedReferences := map[string]model.CardSuggestions{
		"Elemental HERO Sunrise": {
			NamedMaterials:       &[]model.CardReference{},
			MaterialArchetypes:   &[]string{"HERO"},
			NamedReferences:      &[]model.CardReference{{Occurrences: 1, Card: skc_testing.CardMocks["Elemental HERO Sunrise"]}, {Occurrences: 1, Card: skc_testing.CardMocks["Miracle Fusion"]}},
			ReferencedArchetypes: &[]string{"HERO"},
		},
	}

	validateMaterialReferences(
		skc_testing.CardMocks["Elemental HERO Sunrise"],
		*expectedReferences["Elemental HERO Sunrise"].NamedMaterials,
		*expectedReferences["Elemental HERO Sunrise"].MaterialArchetypes,
		assert,
	)
	validateReferences(skc_testing.CardMocks["Elemental HERO Sunrise"],
		*expectedReferences["Elemental HERO Sunrise"].NamedReferences,
		*expectedReferences["Elemental HERO Sunrise"].ReferencedArchetypes,
		assert,
	)
}
