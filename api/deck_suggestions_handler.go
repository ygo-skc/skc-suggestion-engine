package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func getSuggestedDecks(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]
	log.Printf("Getting decks that use card w/ ID: %s", cardID)

	suggestedDecks := model.SuggestedDecks{}

	suggestedDecks.FeaturedIn, _ = skcSuggestionEngineDBInterface.GetDecksThatFeatureCards([]string{cardID})

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(suggestedDecks)
}
