package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

var (
	quotedStringRegex            = regexp.MustCompile("\"[^., ].*?\"")
	deckListCardAndQuantityRegex = regexp.MustCompile("[1-3][xX][0-9]{8}")
)

// Handler that will be used by material suggestion endpoint.
// Will retrieve fusion, synchro, etc materials if they are explicitly mentioned by name and their name exists in the DB.
func getSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]
	log.Println("Getting suggestions for card:", cardID)

	if cardToGetSuggestionsFor, err := db.FindDesiredCardInDBUsingID(cardID); err != nil {
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotFound)

		json.NewEncoder(res).Encode(model.APIError{Message: "Cannot find card using ID " + cardID})
	} else {
		var s model.CardSuggestions

		materialString, _ := cardToGetSuggestionsFor.GetPotentialMaterialsAsString()
		s.NamedMaterials = getMaterials(materialString)

		s.Decks, _ = db.GetDecksThatFeatureCards([]string{cardID})

		res.Header().Add("Content-Type", "application/json")
		json.NewEncoder(res).Encode(s)
	}
}

// Uses regex to find all direct references to cards (or potentially archetypes) and searches it in the DB.
// If a direct name reference is found in the DB, then it is returned as a suggestion.
func getMaterials(materialString string) *[]model.Card {
	namedMaterials, _ := isolateReferences(materialString)

	uniqueMaterials := make([]model.Card, 0, len(namedMaterials))
	for _, card := range namedMaterials {
		uniqueMaterials = append(uniqueMaterials, card)
	}

	sort.SliceStable(uniqueMaterials, func(i, j int) bool {
		return uniqueMaterials[i].CardName < uniqueMaterials[j].CardName // sorting alphabetically from a-z
	})

	if len(uniqueMaterials) < 1 {
		return nil
	} else {
		return &uniqueMaterials
	}
}

func isolateReferences(materialString string) (map[string]model.Card, []string) {
	tokens := quotedStringRegex.FindAllString(materialString, -1)

	namedMaterials := map[string]model.Card{}
	var archetypalMaterials []string

	for _, token := range tokens {
		token = strings.ReplaceAll(token, "\"", "")

		if card, err := db.FindDesiredCardInDBUsingName(token); err != nil {
			archetypalMaterials = append(archetypalMaterials, token)
		} else {
			namedMaterials[card.CardID] = card
		}
	}

	log.Printf("Could not find the following in DB: %v. Potentially an archetype?", archetypalMaterials)
	log.Printf("Found %d unique named materials.", len(namedMaterials))
	return namedMaterials, archetypalMaterials
}
