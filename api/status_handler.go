package api

import (
	"context"
	"fmt"
	"net/http"

	json "github.com/goccy/go-json"
	cModel "github.com/ygo-skc/skc-go/common/v2/model"
	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
)

const (
	statusOp = "Status"
)

// Handler for status/health check endpoint of the api.
// Will get status of downstream services as well to help isolate problems.
func getAPIStatusHandler(res http.ResponseWriter, req *http.Request) {
	logger, ctx := cUtil.InitRequest(context.Background(), apiName, statusOp)

	downstreamHealth := []cModel.DownstreamItem{}

	var ygoServiceVersion string
	var skcSuggestionDBVersion string

	// get status on SKC DB by checking the version number. If this operation fails, its save to assume the DB is down.
	if ygoServiceStatus, err := downstream.YGO.HealthService.GetAPIStatus(ctx); err != nil {
		downstreamHealth = append(downstreamHealth, cModel.DownstreamItem{ServiceName: "YGO Service", Status: cModel.Down})
	} else {
		downstreamHealth = append(downstreamHealth, cModel.DownstreamItem{ServiceName: "YGO Service", Status: cModel.Up, Version: ygoServiceStatus.Version})
		ygoServiceVersion = ygoServiceStatus.Version
	}

	// get status on SKC Suggestion DB by checking the version number. If this operation fails, its save to assume the DB is down.
	if dbVersion, err := skcSuggestionEngineDBInterface.GetSKCSuggestionDBVersion(ctx); err != nil {
		downstreamHealth = append(downstreamHealth, cModel.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: cModel.Down})
	} else {
		downstreamHealth = append(downstreamHealth, cModel.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: cModel.Up})
		skcSuggestionDBVersion = dbVersion
	}

	status := cModel.APIHealth{Version: "2.1.8", Downstream: downstreamHealth}

	logger.Info(fmt.Sprintf("API Status Info! SKC DB version: %s, and SKC Suggestion Engine version: %s", ygoServiceVersion, skcSuggestionDBVersion))
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(status)
}
