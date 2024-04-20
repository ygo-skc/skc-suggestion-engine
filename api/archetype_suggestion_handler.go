package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

type archetypeSuggestionHandlers struct {
	fetchArchetypeSuggestionsHandler func(archetypeName string) ([]model.Card, *model.APIError)
	archetypeSuggestionCBHandler     func([]model.Card, *model.ArchetypalSuggestions)
}

var (
	cardNameArchetypeSuggestionHandlers = archetypeSuggestionHandlers{fetchArchetypeSuggestionsHandler: skcDBInterface.GetInArchetypeSupportUsingCardName, archetypeSuggestionCBHandler: func(dbData []model.Card, as *model.ArchetypalSuggestions) {
		as.UsingName = dbData
	}}
	cardTextArchetypeSuggestionHandlers = archetypeSuggestionHandlers{fetchArchetypeSuggestionsHandler: skcDBInterface.GetInArchetypeSupportUsingCardText, archetypeSuggestionCBHandler: func(dbData []model.Card, as *model.ArchetypalSuggestions) {
		as.UsingText = dbData
	}}
	archetypeExclusionHandlers = archetypeSuggestionHandlers{fetchArchetypeSuggestionsHandler: skcDBInterface.GetArchetypeExclusionsUsingCardText, archetypeSuggestionCBHandler: func(dbData []model.Card, as *model.ArchetypalSuggestions) {
		as.Exclusions = dbData
	}}
)

func getArchetypeSupportHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	archetypeName := pathVars["archetypeName"]
	log.Printf("Getting cards belonging to archetype: %s", archetypeName)

	if err := validation.V.Var(archetypeName, validation.ArchetypeValidator); err != nil {
		log.Printf("%s failed archetype validation", archetypeName)
		validationErr := validation.HandleValidationErrors(err.(validator.ValidationErrors))
		validationErr.HandleServerResponse(res)
		return
	}

	if isBlackListed, err := skcSuggestionEngineDBInterface.IsBlackListed("archetype", archetypeName); err != nil {
		err.HandleServerResponse(res)
		return
	} else if isBlackListed {
		err := model.APIError{Message: fmt.Sprintf("%s is a blacklisted archetype. Common english words are blacklisted. This is done to prevent queries that make no logical sense.", archetypeName), StatusCode: http.StatusBadRequest}
		err.HandleServerResponse(res)
		return
	}

	archetypalSuggestions := model.ArchetypalSuggestions{}

	// setup channels
	supportUsingCardNameChannel, supportUsingTextChannel, exclusionsChannel := make(chan *model.APIError), make(chan *model.APIError), make(chan *model.APIError)

	go getArchetypeSuggestion(archetypeName, &archetypalSuggestions, supportUsingCardNameChannel, cardNameArchetypeSuggestionHandlers)
	go getArchetypeSuggestion(archetypeName, &archetypalSuggestions, supportUsingTextChannel, cardTextArchetypeSuggestionHandlers)
	go getArchetypeSuggestion(archetypeName, &archetypalSuggestions, exclusionsChannel, archetypeExclusionHandlers)

	if err1, err2, err3 := <-supportUsingCardNameChannel, <-supportUsingTextChannel, <-exclusionsChannel; err1 != nil {
		err1.HandleServerResponse(res)
		return
	} else if err2 != nil {
		err2.HandleServerResponse(res)
		return
	} else if err3 != nil {
		err3.HandleServerResponse(res)
		return
	} else if len(archetypalSuggestions.UsingName) < 2 {
		notAnArchetypeErr := model.APIError{Message: "There are fewer than 2 cards matching input string, as such it is likely this phrase is not an archetype.", StatusCode: http.StatusNotFound}
		res.WriteHeader(notAnArchetypeErr.StatusCode)
		json.NewEncoder(res).Encode(notAnArchetypeErr)
		return
	}

	removeExclusions(&archetypalSuggestions)
	archetypalSuggestions.Total = len(archetypalSuggestions.UsingName) + len(archetypalSuggestions.UsingText)

	log.Printf("Returning the following cards related to %s archetype: %d cards found using card names, %d cards found using card text, and excluding %d. cards", archetypeName,
		len(archetypalSuggestions.UsingName), len(archetypalSuggestions.UsingText), len(archetypalSuggestions.UsingText))
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(archetypalSuggestions)
}

func getArchetypeSuggestion(archetypeName string, as *model.ArchetypalSuggestions, c chan *model.APIError, handlers archetypeSuggestionHandlers) {
	if dbData, err := handlers.fetchArchetypeSuggestionsHandler(archetypeName); err != nil {
		c <- err
	} else {
		handlers.archetypeSuggestionCBHandler(dbData, as)
		c <- nil
	}
}

// TODO: add method level documentation, use better variable names, add more inline comments
func removeExclusions(archetypalSuggestions *model.ArchetypalSuggestions) {
	if len(archetypalSuggestions.Exclusions) == 0 {
		return
	}

	// setting up a map of unique exclusions - should prevent multiple traversing of the same list - effectively making the method O(2n)
	uniqueExclusions := map[string]bool{}
	for _, uniqueExclusion := range archetypalSuggestions.Exclusions {
		uniqueExclusions[uniqueExclusion.CardName] = true
		log.Printf("Removing %s as it is explicitly mentioned as not being part of the archetype ", uniqueExclusion.CardName)
	}

	newList := []model.Card{}
	for _, suggestion := range archetypalSuggestions.UsingName {
		if _, isKey := uniqueExclusions[suggestion.CardName]; !isKey {
			newList = append(newList, suggestion)
		}
	}

	archetypalSuggestions.UsingName = newList
}
