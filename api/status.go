package api

import (
	"encoding/json"
	"net/http"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func getStatusHandler(res http.ResponseWriter, req *http.Request) {
	status := model.Status{Version: "1.0.0"}

	res.Header().Add("Content-Type", "application/json")

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(status)
}
