package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func getArchetypeSupportHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	archetypeName := pathVars["archetypeName"]
	log.Printf("Getting cards belonging to archetype: %s", archetypeName)

	archetypalSupport := model.ArchetypalSuggestions{}

	// setup channels
	supportUsingCardNameChannel, supportUsingTextChannel, exclusionsChannel := make(chan *model.APIError), make(chan *model.APIError), make(chan *model.APIError)

	go getArchetypeSuggestionsUsingCardName(archetypeName, &archetypalSupport, supportUsingCardNameChannel)
	go getArchetypeSuggestionsUsingCardText(archetypeName, &archetypalSupport, supportUsingTextChannel)
	go getArchetypeExclusions(archetypeName, &archetypalSupport, exclusionsChannel)

	if err1, err2, err3 := <-supportUsingCardNameChannel, <-supportUsingTextChannel, <-exclusionsChannel; err1 != nil {
		res.WriteHeader(err1.StatusCode)
		json.NewEncoder(res).Encode(err1)
		return
	} else if err2 != nil {
		res.WriteHeader(err1.StatusCode)
		json.NewEncoder(res).Encode(err1)
		return
	} else if err3 != nil {
		res.WriteHeader(err1.StatusCode)
		json.NewEncoder(res).Encode(err1)
		return
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(archetypalSupport)
}

func getArchetypeSuggestionsUsingCardName(archetypeName string, archetypalSupport *model.ArchetypalSuggestions, c chan *model.APIError) {
	if inArchetype, err := skcDBInterface.FindInArchetypeSupportUsingCardName(archetypeName); err != nil {
		c <- err
	} else {
		archetypalSupport.UsingName = inArchetype
		c <- nil
	}
}

func getArchetypeSuggestionsUsingCardText(archetypeName string, archetypalSupport *model.ArchetypalSuggestions, c chan *model.APIError) {
	if inArchetype, err := skcDBInterface.FindInArchetypeSupportUsingCardText(archetypeName); err != nil {
		c <- err
	} else {
		archetypalSupport.UsingText = inArchetype
		c <- nil
	}
}

func getArchetypeExclusions(archetypeName string, archetypalSupport *model.ArchetypalSuggestions, c chan *model.APIError) {
	if exclusions, err := skcDBInterface.FindArchetypeExclusionsUsingCardText(archetypeName); err != nil {
		c <- err
	} else {
		archetypalSupport.Exclusions = exclusions
		c <- nil
	}
}
