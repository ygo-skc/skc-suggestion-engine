package api

import (
	"log"
	"net/http"
)

func SetupMultiplexer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/suggestions/materials", GetMaterialSuggestionsHandler)

	if err := http.ListenAndServe("localhost:9000", mux); err != nil {
		log.Fatalln("There was an error starting server: ", err)
	}
}
