package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	cUtil "github.com/ygo-skc/skc-go/common/v3/util"
)

const (
	cardMechanicsOp = "Card Mechanics"
)

func getCardMechanicsHandler(res http.ResponseWriter, req *http.Request) {
	cardID := chi.URLParam(req, "cardID")

	logger, ctx := cUtil.InitRequest(req.Context(), apiName, cardMechanicsOp, slog.String("card_id", cardID))
	logger.Info("Card mechanics requested")

	cardMechanics, err := skcSuggestionEngineDBInterface.GetCardMechanics(ctx, cardID)
	if err != nil {
		err.HandleServerResponse(res)
		return
	}

	logger.Info("Returning card mechanics",
		slog.Int("effects", len(cardMechanics.Effects)),
		slog.Int("summon_conditions", len(cardMechanics.SummonConditions)))

	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(cardMechanics); err != nil {
		logger.Error("Could not encode card mechanics", slog.Any("err", err))
	}
}
