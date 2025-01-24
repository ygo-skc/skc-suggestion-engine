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
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

type archetypeSuggestionHandlers struct {
	fetchArchetypeSuggestionsHandler func(ctx context.Context, archetypeName string) ([]cModel.Card, *cModel.APIError)
	archetypeSuggestionCBHandler     func([]cModel.Card, *model.ArchetypalSuggestions)
}

var (
	cardNameArchetypeSuggestionHandlers = archetypeSuggestionHandlers{fetchArchetypeSuggestionsHandler: skcDBInterface.GetInArchetypeSupportUsingCardName, archetypeSuggestionCBHandler: func(dbData []cModel.Card, as *model.ArchetypalSuggestions) {
		as.UsingName = dbData
	}}
	cardTextArchetypeSuggestionHandlers = archetypeSuggestionHandlers{fetchArchetypeSuggestionsHandler: skcDBInterface.GetInArchetypeSupportUsingCardText, archetypeSuggestionCBHandler: func(dbData []cModel.Card, as *model.ArchetypalSuggestions) {
		as.UsingText = dbData
	}}
	archetypeExclusionHandlers = archetypeSuggestionHandlers{fetchArchetypeSuggestionsHandler: skcDBInterface.GetArchetypeExclusionsUsingCardText, archetypeSuggestionCBHandler: func(dbData []cModel.Card, as *model.ArchetypalSuggestions) {
		as.Exclusions = dbData
	}}
)

func getArchetypeSupportHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	archetypeName := pathVars["archetypeName"]

	logger, ctx := cUtil.NewRequestSetup(context.Background(), "archetype support", slog.String("archetypeName", archetypeName))
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

	archetypalSuggestions := model.ArchetypalSuggestions{}

	// setup channels
	supportUsingCardNameChannel, supportUsingTextChannel, exclusionsChannel := make(chan *cModel.APIError), make(chan *cModel.APIError), make(chan *cModel.APIError)

	go getArchetypeSuggestion(ctx, archetypeName, &archetypalSuggestions, supportUsingCardNameChannel, cardNameArchetypeSuggestionHandlers)
	go getArchetypeSuggestion(ctx, archetypeName, &archetypalSuggestions, supportUsingTextChannel, cardTextArchetypeSuggestionHandlers)
	go getArchetypeSuggestion(ctx, archetypeName, &archetypalSuggestions, exclusionsChannel, archetypeExclusionHandlers)

	for i := 0; i < 3; i++ {
		select {
		case err := <-supportUsingCardNameChannel:
			if err != nil {
				err.HandleServerResponse(res)
				return
			} else if len(archetypalSuggestions.UsingName) < 2 {
				notAnArchetypeErr := cModel.APIError{
					Message:    fmt.Sprintf("There are fewer than 2 cards matching requested archetype, as such it is likely '%s' is not an archetype. Note: archetypes are case sensitive (eg HERO != Hero).", archetypeName),
					StatusCode: http.StatusNotFound}
				res.WriteHeader(notAnArchetypeErr.StatusCode)
				json.NewEncoder(res).Encode(notAnArchetypeErr)
				return
			}
		case err := <-supportUsingTextChannel:
			if err != nil {
				err.HandleServerResponse(res)
				return
			}
		case err := <-exclusionsChannel:
			if err != nil {
				err.HandleServerResponse(res)
				return
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

func getArchetypeSuggestion(ctx context.Context, archetypeName string, as *model.ArchetypalSuggestions, c chan<- *cModel.APIError, handlers archetypeSuggestionHandlers) {
	if dbData, err := handlers.fetchArchetypeSuggestionsHandler(ctx, archetypeName); err != nil {
		c <- err
	} else {
		handlers.archetypeSuggestionCBHandler(dbData, as)
		c <- nil
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
		uniqueExclusions[uniqueExclusion.CardName] = struct{}{}
		cUtil.LoggerFromContext(ctx).Warn(fmt.Sprintf("Removing %s as it is explicitly mentioned as not being part of the archetype ", uniqueExclusion.CardName))
	}

	newList := []cModel.Card{}
	for _, suggestion := range archetypalSuggestions.UsingName {
		if _, isKey := uniqueExclusions[suggestion.CardName]; !isKey {
			newList = append(newList, suggestion)
		}
	}

	archetypalSuggestions.UsingName = newList
}
