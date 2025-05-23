package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	productCardSuggestionOp = "Product Card Suggestions"
)

func getProductSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	productID := pathVars["productID"]

	logger, ctx := cUtil.NewRequestSetup(cUtil.ContextWithMetadata(context.Background(), apiName, productCardSuggestionOp),
		productCardSuggestionOp, slog.String("product_id", productID))
	logger.Info("Getting product card suggestions")

	cardsInProductChan := make(chan cModel.BatchCardData[cModel.CardIDs])
	go func() {
		cardsInProduct, _ := skcDBInterface.GetCardsFoundInProduct(ctx, productID)
		cardsInProduct.UnknownResources = make(cModel.CardIDs, 0) // by default, no unknown ids
		cardsInProductChan <- cardsInProduct
	}()

	ccIDs, _ := downstream.YGOService.GetCardColorsProto(ctx) // retrieve card color IDs

	var suggestions model.BatchCardSuggestions[cModel.CardIDs]
	var support model.BatchCardSupport[cModel.CardIDs]
	var wg sync.WaitGroup

	wg.Add(2)
	cardsInProduct := <-cardsInProductChan
	go func() { defer wg.Done(); suggestions = getBatchSuggestions(ctx, cardsInProduct, ccIDs.Values) }()
	go func() { defer wg.Done(); support = getBatchSupport(ctx, cardsInProduct) }()
	wg.Wait()

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(model.ProductSuggestions[cModel.CardIDs]{Suggestions: suggestions, Support: support})
}
