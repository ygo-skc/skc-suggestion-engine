package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
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

func getCardSupport(ctx context.Context, subject model.Card) (model.CardSupport, *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	support := model.CardSupport{Card: subject, ReferencedBy: []model.Card{}, MaterialFor: []model.Card{}}
	var s []model.Card
	var err *model.APIError

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
func determineSupportCards(subject model.Card, references []model.Card) ([]model.Card, []model.Card) {
	referencedBy := []model.Card{}
	materialFor := []model.Card{}

	for _, reference := range references {
		materialString := reference.GetPotentialMaterialsAsString()
		materialStringTokens := quotedStringRegex.FindAllString(materialString, -1)

		remainingEffect := strings.Replace(reference.CardEffect, materialString, "", -1) // effect without materials
		remainingEffectTokens := quotedStringRegex.FindAllString(remainingEffect, -1)

		if reference.IsExtraDeckMonster() && subject.IsCardNameInTokens(materialStringTokens) {
			materialFor = append(materialFor, reference)
		}

		if subject.IsCardNameInTokens(remainingEffectTokens) {
			referencedBy = append(referencedBy, reference)
		}
	}

	return referencedBy, materialFor
}

func getBatchSupportHandler(res http.ResponseWriter, req *http.Request) {
	logger, ctx := util.NewRequestSetup(context.Background(), "batch card support")
	logger.Info("Batch card support requested")

	if reqBody := batchRequestValidator(ctx, res, req, noBatchSuggestions, "support"); reqBody == nil {
		return
	} else if suggestionSubjectsCardData, err := skcDBInterface.GetDesiredCardInDBUsingMultipleCardIDs(ctx, reqBody.CardIDs); err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		numValidIDs := len(reqBody.CardIDs) - len(suggestionSubjectsCardData.UnknownResources)
		referencedBy, materialFor := make([]model.Card, 0, 5), make([]model.Card, 0, 5)

		supportChan := make(chan model.CardSupport, numValidIDs)
		go fetchSupportForBatch(ctx, suggestionSubjectsCardData, supportChan)

		for s := range supportChan {
			referencedBy = append(referencedBy, s.ReferencedBy...)
			materialFor = append(materialFor, s.MaterialFor...)
		}

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(model.BatchCardSupport[model.CardIDs]{ReferencedBy: referencedBy, MaterialFor: materialFor, UnknownResources: suggestionSubjectsCardData.UnknownResources})
	}
}

func fetchSupportForBatch(ctx context.Context, suggestionSubjectsCardData model.BatchCardData[model.CardIDs], supportChan chan<- model.CardSupport) {
	uniqueRequestedIDs := make(map[string]bool)
	var wg sync.WaitGroup

	for _, cardInfo := range suggestionSubjectsCardData.CardInfo {
		// card ID is processed or invalid
		if _, exists := uniqueRequestedIDs[cardInfo.CardID]; exists || slices.Contains(suggestionSubjectsCardData.UnknownResources, cardInfo.CardID) {
			continue
		}

		uniqueRequestedIDs[cardInfo.CardID] = true

		wg.Add(1)
		go func(cardInfo model.Card) {
			defer wg.Done()
			y, _ := getCardSupport(ctx, cardInfo)
			supportChan <- y
		}(cardInfo)
	}

	wg.Wait()
	close(supportChan)
}
