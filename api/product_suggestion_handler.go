package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func getProductSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	productID := pathVars["productID"]
	log.Printf("Getting card suggestions for product w/ ID: %s", productID)

	x, _ := skcDBInterface.GetCardsFoundInProduct(productID)
	ccIds, _ := skcDBInterface.GetCardColorIDs() // retrieve card color IDs

	suggestions := getBatchSuggestions(&x.CardInfo, make([]string, 0), ccIds)

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(suggestions)
}
