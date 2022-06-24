package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/contracts"
	"github.com/ygo-skc/skc-suggestion-engine/db"
)

func GetMaterialSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]

	cards := []contracts.Card{
		db.FindDesiredCardInDB(cardID),
		{CardName: "Elemental HERO Avian"},
		{CardName: "Elemental HERO Burstinatrix"},
	}

	res.Header().Add("Content-Type", "application/json")
	json.NewEncoder(res).Encode(cards)
}
