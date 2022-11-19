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

// Handler that will be used by suggestion endpoint.
// Will retrieve fusion, synchro, etc materials and other references if they are explicitly mentioned by name and their name exists in the DB.
func getSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]
	log.Println("Getting suggestions for card:", cardID)

	if cardToGetSuggestionsFor, err := db.FindDesiredCardInDBUsingID(cardID); err != nil {
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotFound)

		json.NewEncoder(res).Encode(err)
	} else {
		var suggestions model.CardSuggestions
		var materialString string

		// get materials if card is from extra deck
		if cardToGetSuggestionsFor.IsExtraDeckMonster() {
			materialString = cardToGetSuggestionsFor.GetPotentialMaterialsAsString()
			suggestions.NamedMaterials = getReferences(materialString)
		}

		// get named references - excludes materials
		suggestions.NamedReferences = getReferences(strings.ReplaceAll(cardToGetSuggestionsFor.CardEffect, materialString, ""))

		// get decks that feature card
		suggestions.Decks, _ = db.GetDecksThatFeatureCards([]string{cardID})

		res.Header().Add("Content-Type", "application/json")
		json.NewEncoder(res).Encode(suggestions)
	}
}

// Uses regex to find all direct references to cards (or potentially archetypes) and searches it in the DB.
// If a direct name reference is found in the DB, then it is returned as a suggestion.
func getReferences(s string) *[]model.CardReference {
	namedReferences, referenceOccurrence, _ := isolateReferences(s)

	uniqueReferences := make([]model.CardReference, 0, len(namedReferences))
	for _, card := range namedReferences {
		uniqueReferences = append(uniqueReferences, model.CardReference{Card: card, Occurrences: referenceOccurrence[card.CardID]})
	}

	sort.SliceStable(uniqueReferences, func(i, j int) bool {
		return uniqueReferences[i].Card.CardName < uniqueReferences[j].Card.CardName // sorting alphabetically from a-z
	})

	return &uniqueReferences
}

func isolateReferences(s string) (map[string]model.Card, map[string]int, []string) {
	tokens := quotedStringRegex.FindAllString(s, -1)

	namedReferences := map[string]model.Card{}
	referenceOccurrence := map[string]int{}
	var archetypalReferences []string

	for _, token := range tokens {
		token = strings.ReplaceAll(token, "\"", "")

		if card, err := db.FindDesiredCardInDBUsingName(token); err != nil {
			archetypalReferences = append(archetypalReferences, token)
		} else {
			namedReferences[card.CardID] = card

			if _, isPresent := referenceOccurrence[card.CardID]; !isPresent {
				referenceOccurrence[card.CardID] = 0
			}
			referenceOccurrence[card.CardID] += 1
		}
	}

	if len(archetypalReferences) > 0 {
		log.Printf("Could not find the following in DB: %v. Potentially an archetype?", archetypalReferences)
	}

	log.Printf("Found %d unique named references.", len(namedReferences))
	return namedReferences, referenceOccurrence, archetypalReferences
}
