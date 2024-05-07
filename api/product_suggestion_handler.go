package api

import (
	"encoding/json"
	"net/http"
)

func getProductSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	productID := "LEDE"
	x, _ := skcDBInterface.GetCardsFoundInProduct(productID)

	ccIds, _ := skcDBInterface.GetCardColorIDs() // retrieve card color IDs
	suggestions := getBatchSuggestions(&x.CardInfo, make([]string, 0), ccIds)

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(suggestions)
}
