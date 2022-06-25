package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/contracts"
	"github.com/ygo-skc/skc-suggestion-engine/db"
)

var (
	quotedString = regexp.MustCompile("\"[^., ].*?\"")
)

func GetMaterialSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]

	cardToGetSuggestionsFor, _ := db.FindDesiredCardInDBUsingID(cardID)
	materialString, _ := GetMaterialString(cardToGetSuggestionsFor)
	cards := GetMaterials(materialString)

	res.Header().Add("Content-Type", "application/json")
	json.NewEncoder(res).Encode(cards)
}

func GetMaterialString(card contracts.Card) (string, error) {
	effectTokens := strings.SplitAfter(card.CardEffect, "\n")

	if len(effectTokens) < 2 {
		// TODO: handle error
	}

	return effectTokens[0], nil
}

func GetMaterials(materialString string) []contracts.Card {
	tokens := quotedString.FindAllString(materialString, -1)

	materials := map[string]contracts.Card{}
	for _, token := range tokens {
		token = strings.ReplaceAll(token, "\"", "")
		card := db.FindDesiredCardInDBUsingName(token)
		materials[card.CardID] = card
	}

	// TODO: can this be done better?
	values := make([]contracts.Card, 0, len(materials))
	for _, v := range materials {
		values = append(values, v)
	}
	return values
}
