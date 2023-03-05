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
	log.Printf("Getting cards that mention archetype: %s", archetypeName)

	archetypalSupport := model.ArchetypalSuggestions{}

	if inArchetype, err := skcDBInterface.FindInArchetypeSupport(archetypeName); err != nil {
		res.WriteHeader(err.StatusCode)
		json.NewEncoder(res).Encode(err)
		return
	} else {
		archetypalSupport.InArchetype = inArchetype
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(archetypalSupport)
}
