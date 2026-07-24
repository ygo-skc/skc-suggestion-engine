package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	cModel "github.com/ygo-skc/skc-go/common/v3/model"
	cUtil "github.com/ygo-skc/skc-go/common/v3/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
)

const (
	statusOp = "Status"
)

// Handler for status/health check endpoint of the api.
// Will get status of downstream services as well to help isolate problems.
func getAPIStatusHandler(res http.ResponseWriter, req *http.Request) {
	logger, ctx := cUtil.InitRequest(req.Context(), apiName, statusOp)

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

	status := cModel.APIHealth{Version: "3.1.2", Downstream: downstreamHealth}

	logger.Info("API Status",
		slog.String("ygo_service_status", string(downstreamHealth[0].Status)),
		slog.String("ygo_service_version", ygoServiceVersion),
		slog.String("skc_suggestion_db_status", string(downstreamHealth[1].Status)),
		slog.String("skc_suggestion_db_version", skcSuggestionDBVersion))
	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(status); err != nil {
		logger.Error("Could not encode API status response", slog.Any("err", err))
	}
}
