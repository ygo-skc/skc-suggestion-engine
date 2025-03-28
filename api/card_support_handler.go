package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	cModel "github.com/ygo-skc/skc-go/common/model"
	"github.com/ygo-skc/skc-go/common/util"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func getCardSupportHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]

	logger, ctx := util.NewRequestSetup(context.Background(), "card support", slog.String("cardID", cardID))
	logger.Info("Getting support cards")

	if cardToGetSupportFor, err := skcDBInterface.GetDesiredCardInDBUsingID(ctx, cardID); err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		if support, err := getCardSupport(ctx, cardToGetSupportFor); err != nil {
			err.HandleServerResponse(res)
			return
		} else {
			res.WriteHeader(http.StatusOK)
			json.NewEncoder(res).Encode(support)
		}
	}
}

func getCardSupport(ctx context.Context, subject cModel.Card) (model.CardSupport, *cModel.APIError) {
	logger := cUtil.LoggerFromContext(ctx)
	support := model.CardSupport{Card: subject, ReferencedBy: []model.CardReference{}, MaterialFor: []model.CardReference{}}
	var s []cModel.Card
	var err *cModel.APIError

	if s, err = skcDBInterface.GetOccurrenceOfCardNameInAllCardEffect(ctx, subject.CardName, subject.CardID); err == nil {
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
func determineSupportCards(subject cModel.Card, references []cModel.Card) ([]model.CardReference, []model.CardReference) {
	referencedBy := []model.CardReference{}
	materialFor := []model.CardReference{}

	for _, reference := range references {
		materialString := reference.GetPotentialMaterialsAsString()
		materialStringTokens := quotedStringRegex.FindAllString(materialString, -1)

		remainingEffect := strings.Replace(reference.CardEffect, materialString, "", -1) // effect without materials
		remainingEffectTokens := quotedStringRegex.FindAllString(remainingEffect, -1)

		if reference.IsExtraDeckMonster() && subject.IsCardNameInTokens(materialStringTokens) {
			materialFor = append(materialFor, model.CardReference{Occurrences: 1, Card: reference})
		}

		if subject.IsCardNameInTokens(remainingEffectTokens) {
			referencedBy = append(referencedBy, model.CardReference{Occurrences: 1, Card: reference})
		}
	}

	return referencedBy, materialFor
}
