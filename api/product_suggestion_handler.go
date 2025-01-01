package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
)

func getProductSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	productID := pathVars["productID"]

	logger, ctx := util.NewRequestSetup(context.Background(), "product card suggestions", slog.String("productID", productID))
	logger.Info("Getting product card suggestions")

	cardsInProductChan := make(chan model.BatchCardData[model.CardIDs])
	go func() {
		cardsInProduct, _ := skcDBInterface.GetCardsFoundInProduct(ctx, productID)
		cardsInProduct.UnknownResources = make(model.CardIDs, 0) // by default, no unknown ids
		cardsInProductChan <- cardsInProduct
	}()

	ccIDs, _ := skcDBInterface.GetCardColorIDs(ctx) // retrieve card color IDs

	var suggestions model.BatchCardSuggestions[model.CardIDs]
	var support model.BatchCardSupport[model.CardIDs]
	var wg sync.WaitGroup

	wg.Add(2)
	cardsInProduct := <-cardsInProductChan
	go func() { defer wg.Done(); suggestions = getBatchSuggestions(ctx, cardsInProduct, ccIDs) }()
	go func() { defer wg.Done(); support = getBatchSupport(ctx, cardsInProduct) }()
	wg.Wait()

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(model.ProductSuggestions[model.CardIDs]{Suggestions: suggestions, Support: support})
}
