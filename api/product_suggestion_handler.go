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
	"github.com/ygo-skc/skc-go/common/ygo"
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

	cards, ccIDs, err := loadPSData(ctx, productID)

	if err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		var suggestions model.BatchCardSuggestions[cModel.CardIDs]
		var support model.BatchCardSupport[cModel.CardIDs]

		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); suggestions = getBatchSuggestions(ctx, *cards, ccIDs.Values) }()
		go func() { defer wg.Done(); support = getBatchSupport(ctx, *cards) }()
		wg.Wait()

		logger.Info("Successfully retrieved product card suggestions")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(model.ProductSuggestions[cModel.CardIDs]{Suggestions: suggestions, Support: support})
	}
}

// load data needed to form product suggestions
func loadPSData(ctx context.Context, productID string) (*cModel.BatchCardData[cModel.CardIDs], *ygo.CardColors, *cModel.APIError) {
	type productCardsRes struct {
		cards *cModel.BatchCardData[cModel.CardIDs]
		err   *cModel.APIError
	}

	var wg sync.WaitGroup
	awg := cUtil.NewAtomicWaitGroup[productCardsRes](&wg)
	go func(awg *cUtil.AtomicWaitGroup[productCardsRes]) {
		productContents, err := downstream.YGO.ProductService.GetCardsByProductIDProto(ctx, productID)
		r := productCardsRes{
			err: err,
		}
		if productContents != nil {
			r.cards = cModel.BatchCardDataFromProductProto[cModel.CardIDs](productContents, cModel.CardIDAsKey)
		}
		awg.Store(&r)
	}(awg)

	ccIDs, err := downstream.YGO.CardService.GetCardColorsProto(ctx)
	if err != nil {
		return nil, nil, err
	}

	r := awg.Load()
	if r.err != nil {
		return nil, nil, r.err
	}

	return r.cards, ccIDs, nil
}
