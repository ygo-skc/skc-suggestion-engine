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
		cardsInProduct, _ := skcDBInterface.GetCardsFoundInProduct(ctx, productID)
		cardsInProduct.UnknownResources = make(model.CardIDs, 0) // by default, no unknown ids
		cardsInProductChan <- cardsInProduct
	}()
	go func() {
		ccIDs, _ := skcDBInterface.GetCardColorIDs(ctx) // retrieve card color IDs
		ccIDsChan <- ccIDs
	}()

	cardsInProduct := <-cardsInProductChan
	suggestions := getBatchSuggestions(ctx, cardsInProduct, <-ccIDsChan)
	support := getBatchSupport(ctx, cardsInProduct)

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(model.ProductSuggestions[model.CardIDs]{Suggestions: suggestions, Support: support})
}
