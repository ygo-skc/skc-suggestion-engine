package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	cModel "github.com/ygo-skc/skc-go/common/v2/model"
	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	similarCardsOp = "Similar Cards"
)

func getSimilarCardsHandler(res http.ResponseWriter, req *http.Request) {
	cardID := chi.URLParam(req, "cardID")

	logger, ctx := cUtil.InitRequest(context.Background(), apiName, similarCardsOp, slog.String("card_id", cardID))
	logger.Info("Finding similar cards")

	subject, embeddedQuery, err := retrieveAndEmbedCardEffect(ctx, cardID)
	if err != nil {
		logger.Error("Could not embed card text", "err", err)
		err.HandleServerResponse(res)
		return
	}

	similarCards := model.SimilarCards{Card: *subject}
	if matches, err := getSimilarCards(ctx, *subject, embeddedQuery); err != nil {
		logger.Error("Could not retrieve similar cards", "err", err)
		err.HandleServerResponse(res)
		return
	} else {
		similarCards.Matches = matches
	}

	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(similarCards); err != nil {
		logger.Error("Could not encode card similarity response", "err", err, "card_id", cardID)
	}
}

func retrieveAndEmbedCardEffect(ctx context.Context, cardID string) (*cModel.YGOCard, []float64, *cModel.APIError) {
	subject, err := downstream.YGO.CardService.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, nil, err
	}

	voyageRes, err := downstream.GetEmbeddings(ctx, []string{(*subject).GetEffect()}, "query")
	if err != nil {
		return nil, nil, err
	}

	return subject, voyageRes.Data[0].Embedding, nil
}

func getSimilarCards(ctx context.Context, subject cModel.YGOCard, embeddedQuery []float64) ([]cModel.YGOCard, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)

	vectorSearchResults, err := skcSuggestionEngineDBInterface.VectorSearchOnCardEmbedding(ctx, subject, embeddedQuery)
	if err != nil {
		return nil, err
	}

	similarCardIDs := make(cModel.CardIDs, 0, len(vectorSearchResults))
	for _, vectorSearchResult := range vectorSearchResults {
		similarCardIDs = append(similarCardIDs, vectorSearchResult.ID)
	}

	similarCardData, err := downstream.YGO.CardService.GetCardsByID(ctx, similarCardIDs)
	if err != nil {
		logger.Error("Could not retrieve data on vector search results", "err", err)
		return nil, err
	}

	if len(similarCardData.UnknownResources) > 0 {
		logger.Warn("Some vector search IDs had no matching metadata", "unknown_card_ids", similarCardData.UnknownResources)
	}

	similarCards := make([]cModel.YGOCard, 0, len(similarCardIDs))
	for _, id := range similarCardIDs {
		if card, isPresent := similarCardData.CardInfo[id]; isPresent {
			similarCards = append(similarCards, card)
		}
	}

	return similarCards, nil
}
