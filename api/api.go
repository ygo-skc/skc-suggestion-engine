// Core package used to configure skc-suggestion-engine api and its endpoints.
package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	CONTEXT = "/api/v1/suggestions"
)

// Configures routes and starts the application server.
func SetupMultiplexer() {
	router := mux.NewRouter()

	router.HandleFunc(CONTEXT+"/status", getStatusHandler)
	router.HandleFunc(CONTEXT+"/materials/{cardID:[0-9]{8}}", getMaterialSuggestionsHandler).Methods(http.MethodGet).Name("Material Suggestion")
	router.HandleFunc(CONTEXT+"/deck", submitNewDeckList).Methods(http.MethodPost).Name("Deck List Submission")
	router.HandleFunc(CONTEXT+"/traffic-analysis", submitNewTrafficData).Methods(http.MethodPost)

	log.Println("Starting server in port 9000")
	if err := http.ListenAndServe(":9000", router); err != nil { // docker does not like localhost:9000 so im just using port number
		log.Fatalln("There was an error starting api server: ", err)
	}
}
