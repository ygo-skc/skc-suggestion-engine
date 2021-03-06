// Core package used to configure skc-suggestion-engine api and its endpoints.
package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ip2location/ip2location-go/v9"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
)

const (
	CONTEXT = "/api/v1/suggestions"
)

var (
	apiKey string
	ipDB   *ip2location.DB
)

func init() {
	// init api key variable
	apiKey = util.EnvMap["API_KEY"]

	// init IP DB
	if ip, err := ip2location.OpenDB("./data/IPv4-DB.BIN"); err != nil {
		log.Fatalln("Could not load IP DB file...")
	} else {
		ipDB = ip
	}
}

func verifyApiKey(headers http.Header) *model.APIError {
	key := headers.Get("API-Key")

	if key != apiKey {
		log.Println("Client is using incorrect API Key.")
		return &model.APIError{Message: "Request has incorrect or missing API Key."}
	}

	return nil
}

// Configures routes and starts the application server.
func SetupMultiplexer() {
	router := mux.NewRouter()

	router.HandleFunc(CONTEXT+"/status", getStatusHandler)
	router.HandleFunc(CONTEXT+"/materials/{cardID:[0-9]{8}}", getMaterialSuggestionsHandler).Methods(http.MethodGet).Name("Material Suggestion")

	router.HandleFunc(CONTEXT+"/deck", submitNewDeckList).Methods(http.MethodPost).Name("Deck List Submission")
	router.HandleFunc(CONTEXT+"/deck/{deckID:[0-9a-z]+}", getDeckList).Methods(http.MethodGet).Name("Retrieve Info On Deck")

	router.HandleFunc(CONTEXT+"/traffic-analysis", submitNewTrafficData).Methods(http.MethodPost).Name("Traffic Analysis")

	log.Println("Starting server in port 9000")
	if err := http.ListenAndServe(":9000", router); err != nil { // docker does not like localhost:9000 so im just using port number
		log.Fatalln("There was an error starting api server: ", err)
	}
}
