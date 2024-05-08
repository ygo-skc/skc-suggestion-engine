package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func getProductSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	productID := pathVars["productID"]
	log.Printf("Getting card suggestions for product w/ ID: %s", productID)

	cardsInProductChan, ccIDsChan := make(chan model.BatchCardData[model.CardIDs]), make(chan map[string]int)
	go func() {
		cardsInProduct, _ := skcDBInterface.GetCardsFoundInProduct(productID)
		cardsInProductChan <- cardsInProduct
	}()
	go func() {
		ccIDs, _ := skcDBInterface.GetCardColorIDs() // retrieve card color IDs
		ccIDsChan <- ccIDs
	}()

	cardsInProduct := <-cardsInProductChan
	suggestions := getBatchSuggestions(&cardsInProduct.CardInfo, make([]string, 0), <-ccIDsChan)

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(suggestions)
}
