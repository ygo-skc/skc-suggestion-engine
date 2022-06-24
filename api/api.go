package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupMultiplexer() {
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/suggestions/materials/{cardID:[0-9]{8}}", GetMaterialSuggestionsHandler).Methods(http.MethodGet)

	if err := http.ListenAndServe("localhost:9000", router); err != nil {
		log.Fatalln("There was an error starting server: ", err)
	}
}
