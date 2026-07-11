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

	subject, err := downstream.YGO.CardService.GetCardByID(ctx, cardID)
	if err != nil {
		err.HandleServerResponse(res)
		return
	}
	similarCards := model.CardSimilarity{Card: *subject}

	vectorSearchResults, err := skcSuggestionEngineDBInterface.GetSimilarCards(ctx, *subject)
	if err != nil {
		err.HandleServerResponse(res)
		return
	}

	similarIDList := make([]string, len(vectorSearchResults))
	for i, card := range vectorSearchResults {
		// if card.ID == cardID {	TODO:
		// 	continue
		// }
		similarIDList[i] = card.ID
	}

	vectorSearchResultsMetadata, _ := downstream.YGO.CardService.GetCardsByID(ctx, similarIDList) // TODO: handle error
	similarCards.Similar = make([]cModel.YGOCard, len(vectorSearchResults))
	for i, id := range similarIDList {
		similarCards.Similar[i] = vectorSearchResultsMetadata.CardInfo[id]
	}

	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(similarCards); err != nil {
		logger.Error("Could not encode card similarity response", "err", err, "card_id", cardID)
	}
}
