package api

import (
	"encoding/json"
	"net/http"

	"github.com/ygo-skc/skc-suggestion-engine/contracts"
)

func GetMaterialSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	cards := []contracts.Card{
		{CardName: "Elemental HERO Avian"},
		{CardName: "Elemental HERO Burstinatrix"},
	}

	res.Header().Add("Content-Type", "application/json")
	json.NewEncoder(res).Encode(cards)
}
