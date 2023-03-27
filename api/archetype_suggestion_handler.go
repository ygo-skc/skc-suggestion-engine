package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

func getArchetypeSupportHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	archetypeName := pathVars["archetypeName"]
	log.Printf("Getting cards belonging to archetype: %s", archetypeName)

	if err := validation.V.Var(archetypeName, validation.ArchetypeValidator); err != nil {
		log.Printf("%s failed archetype validation", archetypeName)
		validationErr := validation.HandleValidationErrors(err.(validator.ValidationErrors))

		// TODO: update status code
		json.NewEncoder(res).Encode(validationErr)
		return
	}

	archetypalSuggestions := model.ArchetypalSuggestions{}

	// setup channels
	supportUsingCardNameChannel, supportUsingTextChannel, exclusionsChannel := make(chan *model.APIError), make(chan *model.APIError), make(chan *model.APIError)

	go getArchetypeSuggestionsUsingCardName(archetypeName, &archetypalSuggestions, supportUsingCardNameChannel)
	go getArchetypeSuggestionsUsingCardText(archetypeName, &archetypalSuggestions, supportUsingTextChannel)
	go getArchetypeExclusions(archetypeName, &archetypalSuggestions, exclusionsChannel)

	if err1, err2, err3 := <-supportUsingCardNameChannel, <-supportUsingTextChannel, <-exclusionsChannel; err1 != nil {
		res.WriteHeader(err1.StatusCode)
		json.NewEncoder(res).Encode(err1)
		return
	} else if err2 != nil {
		res.WriteHeader(err2.StatusCode)
		json.NewEncoder(res).Encode(err2)
		return
	} else if err3 != nil {
		res.WriteHeader(err3.StatusCode)
		json.NewEncoder(res).Encode(err3)
		return
	}

	removeExclusions(&archetypalSuggestions)

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(archetypalSuggestions)
}

func getArchetypeSuggestionsUsingCardName(archetypeName string, archetypalSuggestions *model.ArchetypalSuggestions, c chan *model.APIError) {
	if inArchetype, err := skcDBInterface.FindInArchetypeSupportUsingCardName(archetypeName); err != nil {
		c <- err
	} else {
		archetypalSuggestions.UsingName = inArchetype
		c <- nil
	}
}

func getArchetypeSuggestionsUsingCardText(archetypeName string, archetypalSuggestions *model.ArchetypalSuggestions, c chan *model.APIError) {
	if inArchetype, err := skcDBInterface.FindInArchetypeSupportUsingCardText(archetypeName); err != nil {
		c <- err
	} else {
		archetypalSuggestions.UsingText = inArchetype
		c <- nil
	}
}

func getArchetypeExclusions(archetypeName string, archetypalSuggestions *model.ArchetypalSuggestions, c chan *model.APIError) {
	if exclusions, err := skcDBInterface.FindArchetypeExclusionsUsingCardText(archetypeName); err != nil {
		c <- err
	} else {
		archetypalSuggestions.Exclusions = exclusions
		c <- nil
	}
}

// TODO: add method level documentation, use better variable names, add more inline comments
func removeExclusions(archetypalSuggestions *model.ArchetypalSuggestions) {
	// setup a map of unique exclusions - should save on multiple traversing
	uniqueExclusions := map[string]bool{}
	for _, uniqueExclusion := range archetypalSuggestions.Exclusions {
		uniqueExclusions[uniqueExclusion.CardName] = true
	}

	newList := []model.Card{}
	for _, suggestion := range archetypalSuggestions.UsingName {
		if _, isKey := uniqueExclusions[suggestion.CardName]; !isKey {
			newList = append(newList, suggestion)
		}
	}

	archetypalSuggestions.UsingName = newList
}
