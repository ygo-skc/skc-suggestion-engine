package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/db"
)

var (
	quotedStringRegex            = regexp.MustCompile("\"[^., ].*?\"")
	deckListCardAndQuantityRegex = regexp.MustCompile("[1-3][xX][0-9]{8}")
)

// Handler that will be used by material suggestion endpoint.
// Will retrieve fusion, synchro, etc materials if they are explicitly mentioned by name and their name exists in the DB.
func GetMaterialSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]
	log.Println("Getting suggested materials for card:", cardID)

	if cardToGetSuggestionsFor, err := db.FindDesiredCardInDBUsingID(cardID); err != nil {
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotFound)

		json.NewEncoder(res).Encode(APIError{Message: "Cannot find card using ID " + cardID})
	} else {
		materialString, _ := GetMaterialString(cardToGetSuggestionsFor)
		cards := GetMaterials(materialString)
		log.Println("Found", len(cards), "unique materials")

		res.Header().Add("Content-Type", "application/json")
		json.NewEncoder(res).Encode(cards)
	}
}

// Uses new line as delimiter to split card effect. Materials are found in the first token.
func GetMaterialString(card db.Card) (string, error) {
	effectTokens := strings.SplitAfter(card.CardEffect, "\n")

	if len(effectTokens) < 2 {
		// TODO: handle error
	}

	return effectTokens[0], nil
}

// Uses regex to find all direct references to cards (or potentially archetypes) and searches it in the DB.
// If a direct name reference is found in the DB, then it is returned as a suggestion.
func GetMaterials(materialString string) []db.Card {
	tokens := quotedStringRegex.FindAllString(materialString, -1)

	materials := map[string]db.Card{}
	for _, token := range tokens {
		token = strings.ReplaceAll(token, "\"", "")

		if card, err := db.FindDesiredCardInDBUsingName(token); err != nil {
			log.Println("Could not find the full name", token, "in DB. Potentially an archetype?")
		} else {
			materials[card.CardID] = card
		}
	}

	// TODO: can this be done better?
	values := make([]db.Card, 0, len(materials))
	for _, v := range materials {
		values = append(values, v)
	}
	return values
}
