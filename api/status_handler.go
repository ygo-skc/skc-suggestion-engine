package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
)

// Handler for status/health check endpoint of the api.
// Will get status of downstream services as well to help isolate problems.
func getAPIStatusHandler(res http.ResponseWriter, req *http.Request) {
	logger, ctx := cUtil.NewRequestSetup(context.Background(), "status")

	downstreamHealth := []cModel.DownstreamItem{}

	var skcDBVersion string
	var skcSuggestionDBVersion string

	// get status on SKC DB by checking the version number. If this operation fails, its save to assume the DB is down.
	if dbVersion, err := skcDBInterface.GetSKCDBVersion(ctx); err != nil {
		downstreamHealth = append(downstreamHealth, cModel.DownstreamItem{ServiceName: "SKC API DB", Status: cModel.Down})
	} else {
		downstreamHealth = append(downstreamHealth, cModel.DownstreamItem{ServiceName: "SKC API DB", Status: cModel.Up})
		skcDBVersion = dbVersion
	}

	// get status on SKC Suggestion DB by checking the version number. If this operation fails, its save to assume the DB is down.
	if dbVersion, err := skcSuggestionEngineDBInterface.GetSKCSuggestionDBVersion(ctx); err != nil {
		downstreamHealth = append(downstreamHealth, cModel.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: cModel.Down})
	} else {
		downstreamHealth = append(downstreamHealth, cModel.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: cModel.Up})
		skcSuggestionDBVersion = dbVersion
	}

	status := cModel.APIHealth{Version: "1.5.7", Downstream: downstreamHealth}

	logger.Info(fmt.Sprintf("API Status Info! SKC DB version: %s, and SKC Suggestion Engine version: %s", skcDBVersion, skcSuggestionDBVersion))
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(status)
}
