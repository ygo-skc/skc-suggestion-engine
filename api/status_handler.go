package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func getStatusHandler(res http.ResponseWriter, req *http.Request) {
	log.Print("Getting API status")

	var skcDB model.DownstreamItem
	if _, err := db.GetVersion(); err != nil {
		skcDB = model.DownstreamItem{ServiceName: "SKC API DB", Status: "Down"}
	} else {
		skcDB = model.DownstreamItem{ServiceName: "SKC API DB", Status: "Up"}
	}

	var skcSuggestionDB model.DownstreamItem
	if _, err := db.GetSkcSuggestionDBVersion(); err != nil {
		skcSuggestionDB = model.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: "Down"}
	} else {
		skcSuggestionDB = model.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: "Up"}
	}

	downstream := []model.DownstreamItem{skcDB, skcSuggestionDB}

	status := model.Status{Version: "1.0.0", Downstream: downstream}

	res.Header().Add("Content-Type", "application/json")

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(status)
}
