// Core package used to configure skc-suggestion-engine api and its endpoints.
package api

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/ip2location/ip2location-go/v9"
	"github.com/rs/cors"
	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
)

const (
	CONTEXT = "/api/v1/suggestions"
)

var (
	ipDB                           *ip2location.DB
	skcDBInterface                 db.SKCDatabaseAccessObject = db.SKCDatabaseAccessObjectImplementation{}
	skcSuggestionEngineDBInterface db.SKCSuggestionEngineDAO  = db.SKCSuggestionEngineDAOImplementation{}
	router                         *mux.Router
	corsOpts                       *cors.Cors
)

func init() {
	// init IP DB
	if os.Getenv("IS_TEST") == "false" {
		log.Println("Loading IP DB...")
		if ip, err := ip2location.OpenDB("./data/IPv4-DB.BIN"); err != nil {
			log.Fatalln("Could not load IP DB file...")
		} else {
			ipDB = ip
		}
	} else {
		log.Println("Not loading up IP DB")
	}
}

func verifyApiKey(headers http.Header) *model.APIError {
	clientKey := headers.Get("API-Key")
	realKey := util.EnvMap["API_KEY"]

	if clientKey != realKey {
		log.Println("Client is using incorrect API Key. Cannot process request.")
		return &model.APIError{Message: "Request has incorrect or missing API Key."}
	}

	return nil
}

// sets common headers for response
func commonHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json")
		res.Header().Add("Cache-Control", "max-age=300")

		next.ServeHTTP(res, req)
	})
}

// Configures routes and CORs
func ConfigureServer() {
	router = mux.NewRouter()

	router.HandleFunc(CONTEXT+"/status", getStatusHandler)
	router.HandleFunc(CONTEXT+"/card/{cardID:[0-9]{8}}", getSuggestionsHandler).Methods(http.MethodGet).Name("Material Suggestion")

	router.HandleFunc(CONTEXT+"/deck", submitNewDeckList).Methods(http.MethodPost).Name("Deck List Submission")
	router.HandleFunc(CONTEXT+"/deck/{deckID:[0-9a-z]+}", getDeckList).Methods(http.MethodGet).Name("Retrieve Info On Deck")

	router.HandleFunc(CONTEXT+"/traffic-analysis", submitNewTrafficData).Methods(http.MethodPost).Name("Traffic Analysis")

	// middleware
	router.Use(commonHeadersMiddleware)

	corsOpts = cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://dev.thesupremekingscastle.com", "https://dev.thesupremekingscastle.com", "https://thesupremekingscastle.com", "https://www.thesupremekingscastle.com"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodOptions,
		},

		AllowedHeaders: []string{
			"*", //or you can your header key values which you are using in your application
		},
	})
}

func ServeTLS() {
	log.Println("Starting server in port 9000 (secured)")
	if err := http.ListenAndServeTLS(":9000", "certs/certificate.crt", "certs/private.key", corsOpts.Handler(router)); err != nil { // docker does not like localhost:9000 so im just using port number
		log.Fatalln("There was an error starting api server: ", err)
	}
}

func ServeUnsecured() {
	log.Println("Starting server in port 90 (unsecured)")
	if err := http.ListenAndServe(":90", corsOpts.Handler(router)); err != nil {
		log.Fatalln("There was an error starting api server: ", err)
	}
}
