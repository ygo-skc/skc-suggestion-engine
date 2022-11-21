package api

import (
	"testing"

	"github.com/ygo-skc/skc-suggestion-engine/testing_init"
)

func TestGetReferences(t *testing.T) {
	skcDBInterface = testing_init.SKCDatabaseAccessObjectMock{}

	refs, _ := getReferences(testing_init.CardMocks["Elemental HERO Sunrise"].GetPotentialMaterialsAsString())

	if len(*refs) != 1 {
		t.Errorf("Number of named references did not match.")
	} else {
		t.Logf("Number of named references matched!")
	}
}
