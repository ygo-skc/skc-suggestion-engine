package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
)

func getProductSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	productID := pathVars["productID"]

	logger, ctx := util.NewRequestSetup(context.Background(), "product card suggestions", slog.String("productID", productID))
	logger.Info("Getting product card suggestions")

	cardsInProductChan, ccIDsChan := make(chan model.BatchCardData[model.CardIDs]), make(chan map[string]int)
	go func() {
		cardsInProduct, _ := skcDBInterface.GetCardsFoundInProduct(productID)
		cardsInProductChan <- cardsInProduct
	}()
	go func() {
		ccIDs, _ := skcDBInterface.GetCardColorIDs(ctx) // retrieve card color IDs
		ccIDsChan <- ccIDs
	}()

	cardsInProduct := <-cardsInProductChan
	suggestions := getBatchSuggestions(context.TODO(), &cardsInProduct.CardInfo, make([]string, 0), <-ccIDsChan)

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(suggestions)
}
