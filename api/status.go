package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

// Handler for status/health check endpoint of the api.
// Will get status of downstream services as well to help isolate problems.
func getStatusHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Getting API status")

	var skcDBStatus model.DownstreamItem
	var skcSuggestionDBStatus model.DownstreamItem

	var err error
	var skcDBVersion string
	var skcSuggestionDBVersion string

	// get status on SKC DB by checking the version number. If this operation fails, its save to assume the DB is down.
	if skcDBVersion, err = skcDBInterface.GetSKCDBVersion(); err != nil {
		skcDBStatus = model.DownstreamItem{ServiceName: "SKC API DB", Status: "Down"}
	} else {
		skcDBStatus = model.DownstreamItem{ServiceName: "SKC API DB", Status: "Up"}
	}

	// get status on SKC Suggestion DB by checking the version number. If this operation fails, its save to assume the DB is down.
	if skcSuggestionDBVersion, err = db.GetSKCSuggestionDBVersion(); err != nil {
		skcSuggestionDBStatus = model.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: "Down"}
	} else {
		skcSuggestionDBStatus = model.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: "Up"}
	}

	downstream := []model.DownstreamItem{skcDBStatus, skcSuggestionDBStatus}

	status := model.Status{Version: "1.0.1", Downstream: downstream}

	log.Printf("SKC DB version: %s, and SKC Suggestion Engine version: %s", skcDBVersion, skcSuggestionDBVersion)
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(status)
}
