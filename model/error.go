package model

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string { return e.Message }

func (e *APIError) HandleServerResponse(res http.ResponseWriter) {
	if e.StatusCode == 0 {
		e.StatusCode = 500 // default error code
	}
	res.WriteHeader(e.StatusCode)
	json.NewEncoder(res).Encode(e)
}

func HandleServerResponse(apiErr APIError, res http.ResponseWriter) {
	apiErr.HandleServerResponse(res)
}
