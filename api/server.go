// Core package used to configure skc-suggestion-engine api and its endpoints.
package api

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ip2location/ip2location-go/v9"
	"github.com/rs/cors"
	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/db"
)

const (
	apiContext = "/api/v1/suggestions"
	apiName    = "skc-suggestion-engine"
)

var (
	ipDB *ip2location.DB

	skcSuggestionEngineDBInterface db.SKCSuggestionEngineDAO = db.SKCSuggestionEngineDAOImplementation{}

	serverAPIKey    string
	chicagoLocation *time.Location
)

func init() {
	// init IP DB
	isCICD := os.Getenv("IS_CICD")
	if isCICD != "true" && !strings.HasSuffix(os.Args[0], ".test") {
		slog.Debug("Loading IP DB...")
		if ip, err := ip2location.OpenDB("./data/IPv4-DB9.BIN"); err != nil {
			log.Fatalln("Could not load IP DB file...")
		} else {
			ipDB = ip
		}
	} else {
		slog.Warn("Not loading IP DB")
	}

	// init Location
	if location, err := time.LoadLocation("America/Chicago"); err != nil {
		log.Fatalf("Could not load Chicago location - err %v", err)
	} else {
		chicagoLocation = location
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func verifyApiKey(headers http.Header) *cModel.APIError {
	clientKey := headers.Get("API-Key")

	if clientKey != serverAPIKey {
		slog.Error("Client is using incorrect API Key. Cannot process request")
		return &cModel.APIError{Message: "Request has incorrect or missing API Key."}
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
func commonResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json")
		res.Header().Add("Cache-Control", "max-age=300")

		// gzip
		if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			res.Header().Set("Content-Encoding", "gzip")
			zip := gzip.NewWriter(res)
			next.ServeHTTP(gzipResponseWriter{Writer: zip, ResponseWriter: res}, req)
			zip.Close()
		} else {
			next.ServeHTTP(res, req)
		}
	})
}

// Configures routes and their middle wares
// This method should be called before the environment is set up as the API Key will be set according to the value found in environment
func RunHttpServer() {
	serverAPIKey = cUtil.EnvMap["API_KEY"] // configure API Key
	router := chi.NewRouter()

	// common middleware
	router.Use(commonResponseMiddleware)

	// configure non-admin routes
	router.Route(apiContext, func(r chi.Router) {
		// unprotected routes
		r.Get("/status", getAPIStatusHandler)
		r.Post("/card-details", getBatchCardInfo)
		r.Get("/card-of-the-day", getCardOfTheDay)

		// suggestions
		r.Get(`/card/{cardID:\d{8}}`, getCardSuggestionsHandler)
		r.Post("/card", getBatchSuggestionsHandler)

		// support
		r.Get(`/card/support/{cardID:\d{8}}`, getCardSupportHandler)
		r.Post("/card/support", getBatchSupportHandler)

		r.Get(`/product/{productID:[0-9A-Z]{3,4}}`, getProductSuggestionsHandler)
		r.Get("/archetype/{archetypeName}", getArchetypeSupportHandler)
		r.Get(`/trending/{resource:(?i)card|product}`, trending)

		// admin routes
		r.With(verifyAPIKeyMiddleware).Post("/traffic-analysis", submitNewTrafficDataHandler)
	})

	// Cors
	corsOpts := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "https://dev.thesupremekingscastle.com", "https://thesupremekingscastle.com", "https://www.thesupremekingscastle.com"},
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

	serveTLS(router, corsOpts)
}

// Configures and starts an HTTPS server with TLS encryption.
// It combines the TLS certificate and CA bundle, and utilizes the private key.
// Finally, it applies CORS middleware.
func serveTLS(router *chi.Mux, corsOpts *cors.Cors) {
	slog.Debug("Starting server in port 9000 (secured)")

	cUtil.CombineCerts("certs")
	if err := http.ListenAndServeTLS(":9000", "certs/concatenated.crt", "certs/private.key", corsOpts.Handler(router)); err != nil {
		log.Fatalf("There was an error starting api server: %s", err)
	}
}
