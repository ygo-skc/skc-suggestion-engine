package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

const (
	archetypeSupportOp = "Archetype Support"
)

type archetypeResults struct {
	cards []cModel.YGOCard
	err   *cModel.APIError
}

func getArchetypeSupportHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	archetypeName := pathVars["archetypeName"]

	logger, ctx := cUtil.NewRequestSetup(
		cUtil.ContextWithMetadata(context.Background(), apiName, archetypeSupportOp),
		archetypeSupportOp, slog.String("archetype_name", archetypeName),
	)
	logger.Info("Getting cards within archetype")

	if err := validation.V.Var(archetypeName, validation.ArchetypeValidator); err != nil {
		logger.Error("Failed archetype validation")
		validationErr := validation.HandleValidationErrors(err.(validator.ValidationErrors))
		validationErr.HandleServerResponse(res)
		return
	}

	if isBlackListed, err := skcSuggestionEngineDBInterface.IsBlackListed(ctx, "archetype", archetypeName); err != nil {
		err.HandleServerResponse(res)
		return
	} else if isBlackListed {
		err := cModel.APIError{Message: fmt.Sprintf("%s is a blacklisted archetype. Common english words are blacklisted. This is done to prevent queries that make no logical sense.", archetypeName), StatusCode: http.StatusUnprocessableEntity}
		err.HandleServerResponse(res)
		return
	}

	// setup channels
	supportUsingCardNameChannel, supportUsingTextChannel, exclusionsChannel := make(chan archetypeResults),
		make(chan archetypeResults), make(chan archetypeResults)

	go getArchetypeSuggestion(ctx, archetypeName, supportUsingCardNameChannel,
		downstream.YGOClient.GetArchetypalCardsUsingCardName)
	go getArchetypeSuggestion(ctx, archetypeName, supportUsingTextChannel,
		downstream.YGOClient.GetExplicitArchetypalInclusions)
	go getArchetypeSuggestion(ctx, archetypeName, exclusionsChannel,
		downstream.YGOClient.GetExplicitArchetypalExclusions)

	archetypalSuggestions := model.ArchetypalSuggestions{}
	for i := 0; i < 3; i++ {
		select {
		case ar := <-supportUsingCardNameChannel:
			if ar.err != nil {
				ar.err.HandleServerResponse(res)
				return
			} else if len(ar.cards) < 2 {
				notAnArchetypeErr := cModel.APIError{
					Message:    fmt.Sprintf("There are fewer than 2 cards matching requested archetype, as such it is likely '%s' is not an archetype. Note: archetypes are case sensitive (eg HERO != Hero).", archetypeName),
					StatusCode: http.StatusNotFound}
				res.WriteHeader(notAnArchetypeErr.StatusCode)
				json.NewEncoder(res).Encode(notAnArchetypeErr)
				return
			} else {
				archetypalSuggestions.UsingName = ar.cards
			}
		case ar := <-supportUsingTextChannel:
			if ar.err != nil {
				ar.err.HandleServerResponse(res)
				return
			} else {
				archetypalSuggestions.UsingText = ar.cards
			}
		case ar := <-exclusionsChannel:
			if ar.err != nil {
				ar.err.HandleServerResponse(res)
				return
			} else {
				archetypalSuggestions.Exclusions = ar.cards
			}
		}
	}

	removeExclusions(ctx, &archetypalSuggestions)
	archetypalSuggestions.Total = len(archetypalSuggestions.UsingName) + len(archetypalSuggestions.UsingText)

	logger.Info(fmt.Sprintf("Returning the following cards related to %s archetype: %d cards found using card names, %d cards found using card text, and excluding %d cards", archetypeName,
		len(archetypalSuggestions.UsingName), len(archetypalSuggestions.UsingText), len(archetypalSuggestions.UsingText)))
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(archetypalSuggestions)
}

func getArchetypeSuggestion(ctx context.Context, archetypeName string, c chan<- archetypeResults,
	fetchSuggestions func(context.Context, string) ([]cModel.YGOCard, *cModel.APIError)) {
	if dbData, err := fetchSuggestions(ctx, archetypeName); err != nil {
		c <- archetypeResults{cards: nil, err: err}
	} else if dbData != nil {
		c <- archetypeResults{cards: dbData, err: nil}
	} else {
		c <- archetypeResults{cards: make([]cModel.YGOCard, 0), err: nil}
	}
}

// TODO: add method level documentation, use better variable names, add more inline comments
func removeExclusions(ctx context.Context, archetypalSuggestions *model.ArchetypalSuggestions) {
	if len(archetypalSuggestions.Exclusions) == 0 {
		return
	}

	// setting up a map of unique exclusions - should prevent multiple traversing of the same list - effectively making the method O(2n)
	uniqueExclusions := make(map[string]struct{})
	for _, uniqueExclusion := range archetypalSuggestions.Exclusions {
		uniqueExclusions[uniqueExclusion.GetName()] = struct{}{}
		cUtil.LoggerFromContext(ctx).Warn(fmt.Sprintf("Removing %s as it is explicitly mentioned as not being part of the archetype ", uniqueExclusion.GetName()))
	}

	newList := []cModel.YGOCard{}
	for _, suggestion := range archetypalSuggestions.UsingName {
		if _, isKey := uniqueExclusions[suggestion.GetName()]; !isKey {
			newList = append(newList, suggestion)
		}
	}

	archetypalSuggestions.UsingName = newList
}
