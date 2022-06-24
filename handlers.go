package main

import (
	"encoding/json"
	"net/http"
)

func GetMaterialSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	cards := []Card{
		{CardName: "Elemental HERO Avian"},
		{CardName: "Elemental HERO Burstinatrix"},
	}

	res.Header().Add("Content-Type", "application/json")
	json.NewEncoder(res).Encode(cards)
}
