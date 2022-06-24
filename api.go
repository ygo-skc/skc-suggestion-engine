package main

import (
	"log"
	"net/http"
)

func SetupMultiplexer() {
	http.HandleFunc("/api/v1/suggestions/materials", GetMaterialSuggestionsHandler)

	if err := http.ListenAndServe("localhost:9000", nil); err != nil {
		log.Fatalln("There was an error starting server: ", err)
	}
}
