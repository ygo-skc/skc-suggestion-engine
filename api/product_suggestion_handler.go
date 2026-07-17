package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	cModel "github.com/ygo-skc/skc-go/common/v2/model"
	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"github.com/ygo-skc/skc-go/common/v2/ygo"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/suggest"
)

const (
	productCardSuggestionOp = "Product Card Suggestions"
)

func getProductSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	productID := chi.URLParam(req, "productID")

	logger, ctx := cUtil.InitRequest(context.Background(), apiName, productCardSuggestionOp,
		slog.String("product_id", productID))
	logger.Info("Getting product card suggestions")

	cards, ccIDs, relevantArchetypes, err := loadPSData(ctx, productID)
	if err != nil {
		logger.Error("Failed to retrieve product data", "err", err)
		err.HandleServerResponse(res)
		return
	}

	var suggestions model.BatchCardSuggestions[cModel.CardIDs]
	var support model.BatchCardSupport[cModel.CardIDs]

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		suggestions = getBatchSuggestions(ctx, *cards, relevantArchetypes, ccIDs.Values)
	}()
	go func() { defer wg.Done(); support = getBatchSupport(ctx, *cards) }()
	wg.Wait()

	logger.Info("Successfully retrieved product card suggestions")
	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(model.ProductSuggestions[cModel.CardIDs]{Suggestions: suggestions, Support: support}); err != nil {
		logger.Error("Could not encode product suggestions response", "err", err, "product_id", productID)
	}
}

// load data needed to form product suggestions
func loadPSData(ctx context.Context,
	productID string) (*cModel.BatchCardData[cModel.CardIDs], *ygo.CardColors, []string, *cModel.APIError) {
	productContents, err := downstream.YGO.ProductService.GetCardsByProductIDProto(ctx, productID)
	if err != nil {
		return nil, nil, nil, err
	}
	cards := cModel.BatchCardDataFromProductProto[cModel.CardIDs](productContents, cModel.CardIDAsKey)

	cardIDs := make(cModel.CardIDs, 0, len(cards.CardInfo))
	for id := range cards.CardInfo {
		cardIDs = append(cardIDs, id)
	}

	ccIDs, relevantArchetypes, err := suggest.FetchMetadata(ctx, cardIDs, skcSuggestionEngineDBInterface)
	if err != nil {
		return nil, nil, nil, err
	}

	return cards, ccIDs, relevantArchetypes, nil
}
