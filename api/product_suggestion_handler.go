package api

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	json "github.com/goccy/go-json"
	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	productCardSuggestionOp = "Product Card Suggestions"
)

func getProductSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	productID := chi.URLParam(req, "productID")

	logger, ctx := cUtil.InitRequest(context.Background(), apiName, productCardSuggestionOp,
		slog.String("product_id", productID))
	logger.Info("Getting product card suggestions")

	var wg sync.WaitGroup
	awg := model.NewAtomicWaitGroup[cModel.BatchCardData[cModel.CardIDs]](&wg)
	go func(awg *model.AtomicWaitGroup[cModel.BatchCardData[cModel.CardIDs]]) {
		productContents, _ := downstream.YGO.ProductService.GetCardsByProductIDProto(ctx, productID)
		awg.Store(cModel.BatchCardDataFromProductProto[cModel.CardIDs](productContents, cModel.CardIDAsKey))
	}(awg)

	ccIDs, _ := downstream.YGO.CardService.GetCardColorsProto(ctx) // retrieve card color IDs

	var suggestions model.BatchCardSuggestions[cModel.CardIDs]
	var support model.BatchCardSupport[cModel.CardIDs]

	cardsInProduct := awg.Load()

	wg.Add(2)
	go func() { defer wg.Done(); suggestions = getBatchSuggestions(ctx, *cardsInProduct, ccIDs.Values) }()
	go func() { defer wg.Done(); support = getBatchSupport(ctx, *cardsInProduct) }()
	wg.Wait()

	logger.Info("Successfully retrieved product card suggestions")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(model.ProductSuggestions[cModel.CardIDs]{Suggestions: suggestions, Support: support})
}
