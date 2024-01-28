// Core package used to configure skc-suggestion-engine api and its endpoints.
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	skcDBInterface                 db.SKCDatabaseAccessObject = db.SKCDAOImplementation{}
	skcSuggestionEngineDBInterface db.SKCSuggestionEngineDAO  = db.SKCSuggestionEngineDAOImplementation{}
	serverAPIKey                   string
	chicagoLocation                *time.Location
)

func init() {
	// init IP DB
	isCICD := os.Getenv("IS_CICD")
	if isCICD != "true" && !strings.HasSuffix(os.Args[0], ".test") {
		log.Println("Loading IP DB...")
		if ip, err := ip2location.OpenDB("./data/IPv4-DB.BIN"); err != nil {
			log.Fatalln("Could not load IP DB file...")
		} else {
			ipDB = ip
		}
	} else {
		log.Println("Not loading IP DB")
	}

	// init Location
	if location, err := time.LoadLocation("America/Chicago"); err != nil {
		log.Fatalln("Could not load Chicago location")
	} else {
		chicagoLocation = location
	}
}

func verifyApiKey(headers http.Header) *model.APIError {
	clientKey := headers.Get("API-Key")

	if clientKey != serverAPIKey {
		log.Println("Client is using incorrect API Key. Cannot process request.")
		return &model.APIError{Message: "Request has incorrect or missing API Key."}
	}

	return nil
}

func verifyAPIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := verifyApiKey(req.Header); err != nil {
			res.Header().Add("Content-Type", "application/json")
			res.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(res).Encode(err)
		} else {
			next.ServeHTTP(res, req)
		}
	})
}

// sets common headers for response
func commonHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json")
		res.Header().Add("Cache-Control", "max-age=300")

		next.ServeHTTP(res, req)
	})
}

// Configures routes and their middle wares
// This method should be called before the environment is set up as the API Key will be set according to the value found in environment
func RunHttpServer() {
	configureEnv()
	router := mux.NewRouter()

	// configure non-admin routes
	unprotectedRoutes := router.PathPrefix(CONTEXT).Subrouter()
	unprotectedRoutes.HandleFunc("/status", getAPIStatusHandler)
	unprotectedRoutes.HandleFunc("/card-details", getBatchCardInfo)
	unprotectedRoutes.HandleFunc("/card-of-the-day", getCardOfTheDay).Methods(http.MethodGet).Name("Card of the Day")
	unprotectedRoutes.HandleFunc("/card/{cardID:[0-9]{8}}", getCardSuggestionsHandler).Methods(http.MethodGet).Name("Material Suggestion")
	unprotectedRoutes.HandleFunc("/card/{cardID:[0-9]{8}}/support", getCardSupportHandler).Methods(http.MethodGet).Name("Card Support Suggestions")
	unprotectedRoutes.HandleFunc("/archetype/{archetypeName}", getArchetypeSupportHandler).Methods(http.MethodGet).Name("Archetype Suggestions")
	unprotectedRoutes.HandleFunc("/trending/{resource:(?i)card|product}", trending).Methods(http.MethodGet).Name("Trending")

	// admin routes
	protectedRoutes := router.PathPrefix(CONTEXT).Subrouter()
	protectedRoutes.Use(verifyAPIKeyMiddleware)
	protectedRoutes.HandleFunc("/traffic-analysis", submitNewTrafficDataHandler).Methods(http.MethodPost).Name("Traffic Analysis")

	// common middleware
	router.Use(commonHeadersMiddleware)

	// Cors
	corsOpts := cors.New(cors.Options{
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

	go serveTLS(router, corsOpts)
	serveUnsecured(router, corsOpts)
}

// configure server to handle HTTPS (secured) calls
func serveTLS(router *mux.Router, corsOpts *cors.Cors) {
	log.Println("Starting server in port 9000 (secured)")
	if err := http.ListenAndServeTLS(":9000", "certs/certificate.crt", "certs/private.key", corsOpts.Handler(router)); err != nil { // docker does not like localhost:9000 so im just using port number
		log.Fatalf("There was an error starting api server: %s", err)
	}
}

// configure server to handle HTTPs (un-secured) calls
func serveUnsecured(router *mux.Router, corsOpts *cors.Cors) {
	log.Println("Starting server in port 90 (unsecured)")
	if err := http.ListenAndServe(":90", corsOpts.Handler(router)); err != nil {
		log.Fatalf("There was an error starting api server: %s", err)
	}
}

func configureEnv() {
	serverAPIKey = util.EnvMap["API_KEY"] // configure API Key
	log.Printf("API Key for API %s", serverAPIKey)
}
