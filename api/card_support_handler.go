package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	json "github.com/goccy/go-json"
	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	cardSupportOp = "Card Support"
)

func getCardSupportHandler(res http.ResponseWriter, req *http.Request) {
	cardID := chi.URLParam(req, "cardID")

	logger, ctx := cUtil.InitRequest(context.Background(), apiName, cardSupportOp, slog.String("card_id", cardID))
	logger.Info("Getting support cards")

	if cardToGetSupportFor, err := downstream.YGO.CardService.GetCardByID(ctx, cardID); err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		if support, err := getCardSupport(ctx, *cardToGetSupportFor); err != nil {
			err.HandleServerResponse(res)
			return
		} else {
			res.WriteHeader(http.StatusOK)
			json.NewEncoder(res).Encode(support)
		}
	}
}

func getCardSupport(ctx context.Context, subject cModel.YGOCard) (model.CardSupport, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	support := model.CardSupport{Card: subject, ReferencedBy: []model.CardReference{}, MaterialFor: []model.CardReference{}}
	var s []cModel.YGOCard
	var err *cModel.APIError

	if s, err = downstream.YGO.CardService.SearchForCardRefUsingEffect(ctx, subject.GetName(), subject.GetID()); err == nil {
		if len(s) == 0 {
			logger.Warn("No support found")
			return support, nil
		} else {
			support.ReferencedBy, support.MaterialFor = determineSupportCards(support.Card, s)
			logger.Info(fmt.Sprintf("%d direct references (excluding cards referencing it as a material)", len(support.ReferencedBy)))
			logger.Info(fmt.Sprintf("Can be used as a material for %d cards", len(support.MaterialFor)))
		}
	}
	return support, err
}

// Iterates over a list of support cards and attempts to determine if subject is found in material clause or within the body of the reference.
// If the name is found in the material clause, we can assume the subject is a required or optional summoning material - otherwise its a support card.
func determineSupportCards(subject cModel.YGOCard, references []cModel.YGOCard) ([]model.CardReference, []model.CardReference) {
	referencedBy := []model.CardReference{}
	materialFor := []model.CardReference{}

	for _, reference := range references {
		materialString := cModel.GetPotentialMaterialsAsString(reference)
		materialStringTokens := quotedStringRegex.FindAllString(materialString, -1)

		remainingEffect := strings.Replace(reference.GetEffect(), materialString, "", -1) // effect without materials
		remainingEffectTokens := quotedStringRegex.FindAllString(remainingEffect, -1)

		if cModel.IsExtraDeckMonster(reference) && cModel.IsCardNameInTokens(subject, materialStringTokens) {
			materialFor = append(materialFor, model.CardReference{Occurrences: 1, Card: reference})
		}

		if cModel.IsCardNameInTokens(subject, remainingEffectTokens) {
			referencedBy = append(referencedBy, model.CardReference{Occurrences: 1, Card: reference})
		}
	}

	return referencedBy, materialFor
}
