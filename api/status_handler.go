package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

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

	downstreamHealth := make([]cModel.DownstreamItem, 2)

	var ygoServiceVersion string
	var skcSuggestionDBVersion string

	var wg sync.WaitGroup
	wg.Add(2)

	// get status on SKC DB by checking the version number. If this operation fails, its save to assume the DB is down.
	go func() {
		defer wg.Done()
		if ygoServiceStatus, err := downstream.YGO.HealthService.GetAPIStatus(ctx); err != nil {
			downstreamHealth[0] = cModel.DownstreamItem{ServiceName: "YGO Service", Status: cModel.Down}
		} else {
			downstreamHealth[0] = cModel.DownstreamItem{ServiceName: "YGO Service", Status: cModel.Up, Version: ygoServiceStatus.Version}
			ygoServiceVersion = ygoServiceStatus.Version
		}
	}()

	// get status on SKC Suggestion DB by checking the version number. If this operation fails, its save to assume the DB is down.
	go func() {
		defer wg.Done()
		if dbVersion, err := skcSuggestionEngineDBInterface.GetSKCSuggestionDBVersion(ctx); err != nil {
			downstreamHealth[1] = cModel.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: cModel.Down}
		} else {
			downstreamHealth[1] = cModel.DownstreamItem{ServiceName: "SKC Suggestion Engine DB", Status: cModel.Up}
			skcSuggestionDBVersion = dbVersion
		}
	}()

	wg.Wait()

	status := cModel.APIHealth{Version: "2.2.5", Downstream: downstreamHealth}

	logger.Info("API Status",
		"ygo_service_status", downstreamHealth[0].Status, "ygo_service_version", ygoServiceVersion,
		"skc_suggestion_db_status", downstreamHealth[1].Status, "skc_suggestion_db_version", skcSuggestionDBVersion)
	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(status); err != nil {
		logger.Error("Could not encode API status response", "err", err)
	}
}
