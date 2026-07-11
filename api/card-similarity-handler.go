package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
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

	err = skcSuggestionEngineDBInterface.GetSimilarCards(ctx, *subject)

	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(subject); err != nil {
		// TODO: add logger
	}
}
