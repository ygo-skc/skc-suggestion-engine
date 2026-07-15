package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	cModel "github.com/ygo-skc/skc-go/common/v2/model"
	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

const (
	archetypeSupportOp   = "Archetype Support"
	archetypeSupportV2Op = "Archetype Support v2"
)

type archetypeResults struct {
	cards []cModel.YGOCard
	err   *cModel.APIError
}

func getArchetypeSupportHandler(res http.ResponseWriter, req *http.Request) {
	archetypeName := chi.URLParam(req, "archetypeName")

	logger, ctx := cUtil.InitRequest(context.Background(), apiName, archetypeSupportOp, slog.String("archetype_name", archetypeName))
	logger.Info("Getting cards within archetype")

	if err := validation.V.Var(archetypeName, validation.ArchetypeValidator); err != nil {
		logger.Error("Failed archetype validation", "err", err)
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
	supportUsingCardNameChannel, supportUsingTextChannel, exclusionsChannel := make(chan archetypeResults, 1),
		make(chan archetypeResults, 1), make(chan archetypeResults, 1)

	go getArchetypeSuggestion(ctx, archetypeName, supportUsingCardNameChannel,
		downstream.YGO.CardService.GetArchetypalCardsUsingCardName)
	go getArchetypeSuggestion(ctx, archetypeName, supportUsingTextChannel,
		downstream.YGO.CardService.GetExplicitArchetypalInclusions)
	go getArchetypeSuggestion(ctx, archetypeName, exclusionsChannel,
		downstream.YGO.CardService.GetExplicitArchetypalExclusions)

	archetypalSuggestions := model.ArchetypalSuggestions{}
	for range 3 {
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
				if err := json.NewEncoder(res).Encode(notAnArchetypeErr); err != nil {
					logger.Error("Could not encode archetype error response", "err", err, "archetype_name", archetypeName)
				}
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

	logger.Info("Returning archetypal suggestions",
		"archetype_name", archetypeName,
		"cards_found_using_name", len(archetypalSuggestions.UsingName),
		"cards_found_using_text", len(archetypalSuggestions.UsingText),
		"excluded_cards", len(archetypalSuggestions.Exclusions))

	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(archetypalSuggestions); err != nil {
		logger.Error("Could not encode archetypal suggestions response", "err", err, "archetype_name", archetypeName, "total_cards", archetypalSuggestions.Total)
	}
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
		cUtil.RetrieveLogger(ctx).Warn("Card explicitly excluded from archetype", "card_name", uniqueExclusion.GetName())
	}

	newList := []cModel.YGOCard{}
	for _, suggestion := range archetypalSuggestions.UsingName {
		if _, isKey := uniqueExclusions[suggestion.GetName()]; !isKey {
			newList = append(newList, suggestion)
		}
	}

	archetypalSuggestions.UsingName = newList
}

func getArchetypeSupportV2Handler(res http.ResponseWriter, req *http.Request) {
	archetypeName := chi.URLParam(req, "archetypeName")

	logger, ctx := cUtil.InitRequest(context.Background(), apiName, archetypeSupportV2Op, slog.String("archetype_name", archetypeName))
	logger.Info("Getting cards within archetype")

	if err := validation.V.Var(archetypeName, validation.ArchetypeValidator); err != nil {
		logger.Error("Failed archetype validation", "err", err)
		validationErr := validation.HandleValidationErrors(err.(validator.ValidationErrors))
		validationErr.HandleServerResponse(res)
		return
	}

	i, q, e, err := skcSuggestionEngineDBInterface.GetArchetypeMembers(ctx, archetypeName)
	if err != nil {
		err.HandleServerResponse(res)
		return
	}

	m := make(cModel.CardIDs, 0, len(i)+len(q)+len(e))
	for _, item := range i {
		m = append(m, item)
	}
	for _, item := range q {
		m = append(m, item)
	}
	for _, item := range e {
		m = append(m, item)
	}

	batchCardInfo, err := downstream.YGO.CardService.GetCardsByID(ctx, m)
	if err != nil {
		err.HandleServerResponse(res)
		return
	}

	archetypeMembers := model.ArchetypeMembers{
		Archetype:        archetypeName,
		InheritMembers:   make([]cModel.YGOCard, 0, len(i)),
		QualifiedMembers: make([]cModel.YGOCard, 0, len(q)),
		ExcludedMembers:  make([]cModel.YGOCard, 0, len(e)),
	}

	for _, member := range i {
		archetypeMembers.InheritMembers = append(archetypeMembers.InheritMembers, batchCardInfo.CardInfo[member])
	}

	for _, member := range q {
		archetypeMembers.QualifiedMembers = append(archetypeMembers.QualifiedMembers, batchCardInfo.CardInfo[member])
	}

	for _, member := range e {
		archetypeMembers.ExcludedMembers = append(archetypeMembers.ExcludedMembers, batchCardInfo.CardInfo[member])
	}

	logger.Info("Returning archetypal suggestions",
		"archetype_name", archetypeName,
		"inherit_members", len(archetypeMembers.InheritMembers),
		"qualified_members", len(archetypeMembers.QualifiedMembers),
		"excluded_members", len(archetypeMembers.ExcludedMembers))

	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(archetypeMembers); err != nil {
		logger.Error("Could not encode archetypal suggestions response", "err", err, "archetype_name", archetypeName)
	}
}
