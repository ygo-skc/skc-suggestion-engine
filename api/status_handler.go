package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

// Handler for status/health check endpoint of the api.
// Will get status of downstream services as well to help isolate problems.
func getAPIStatusHandler(res http.ResponseWriter, req *http.Request) {
	downstreamHealth := []model.DownstreamItem{}

	var skcDBVersion string
	var skcSuggestionDBVersion string

	// get status on SKC DB by checking the version number. If this operation fails, its save to assume the DB is down.
	if dbVersion, err := skcDBInterface.GetSKCDBVersion(); err != nil {
		downstreamHealth = append(downstreamHealth, model.DownstreamItem{ServiceName: "SKC API DB", Status: model.Down})
	} else {
		downstreamHealth = append(downstreamHealth, model.DownstreamItem{ServiceName: "SKC API DB", Status: model.Up})
		skcDBVersion = dbVersion
	}

	// get status on SKC Suggestion DB by checking the version number. If this operation fails, its save to assume the DB is down.
	if dbVersion, err := skcSuggestionEngineDBInterface.GetSKCSuggestionDBVersion(); err != nil {
		downstreamHealth = append(downstreamHealth, model.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: model.Down})
	} else {
		downstreamHealth = append(downstreamHealth, model.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: model.Up})
		skcSuggestionDBVersion = dbVersion
	}

	status := model.APIHealth{Version: "1.4.3", Downstream: downstreamHealth}

	log.Printf("API Status Info! SKC DB version: %s, and SKC Suggestion Engine version: %s", skcDBVersion, skcSuggestionDBVersion)
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(status)
}
