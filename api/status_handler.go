package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
)

// Handler for status/health check endpoint of the api.
// Will get status of downstream services as well to help isolate problems.
func getAPIStatusHandler(res http.ResponseWriter, req *http.Request) {
	logger, ctx := util.NewRequestSetup(context.Background(), "status")

	downstreamHealth := []model.DownstreamItem{}

	var skcDBVersion string
	var skcSuggestionDBVersion string

	// get status on SKC DB by checking the version number. If this operation fails, its save to assume the DB is down.
	if dbVersion, err := skcDBInterface.GetSKCDBVersion(ctx); err != nil {
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

	status := model.APIHealth{Version: "1.4.5", Downstream: downstreamHealth}

	logger.Info(fmt.Sprintf("API Status Info! SKC DB version: %s, and SKC Suggestion Engine version: %s", skcDBVersion, skcSuggestionDBVersion))
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(status)
}
