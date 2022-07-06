// Core package used to configure skc-suggestion-engine api and its endpoints.
package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Configures routes and starts the application server.
func SetupMultiplexer() {
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/suggestions/materials/{cardID:[0-9]{8}}", GetMaterialSuggestionsHandler).Methods(http.MethodGet).Name("Material Suggestion")
	router.HandleFunc("/api/v1/suggestions/deck", SubmitNewDeckList).Methods(http.MethodPost).Queries("list", "{list}").Queries("name", "{name}").Queries("tags", "{tags}").Name("Deck List Submission")
	//

	log.Println("Starting server in port 9000")
	if err := http.ListenAndServe(":9000", router); err != nil { // docker does not like localhost:9000 so im just using port number
		log.Fatalln("There was an error starting api server: ", err)
	}
}
