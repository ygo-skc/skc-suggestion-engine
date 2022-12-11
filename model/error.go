package model

type APIError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string { return e.Message }
