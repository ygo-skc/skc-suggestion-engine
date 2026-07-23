package api

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ip2location/ip2location-go/v9"
	"github.com/rs/cors"
	cModel "github.com/ygo-skc/skc-go/common/v3/model"
	cUtil "github.com/ygo-skc/skc-go/common/v3/util"
	"github.com/ygo-skc/skc-suggestion-engine/db"
	"golang.org/x/net/http2"
)

const (
	v1Context = "/api/v1/suggestions"
	v2Context = "/api/v2/suggestions"
	apiName   = "skc-suggestion-engine"
	apiPort   = 9000
)

var (
	ipDB *ip2location.DB

	skcSuggestionEngineDBInterface db.SKCSuggestionEngineDAO = db.SKCSuggestionEngineDAOImplementation{}

	serverAPIKey    string
	chicagoLocation *time.Location

	gzipPool = sync.Pool{
		New: func() any {
			w, _ := gzip.NewWriterLevel(io.Discard, 2)
			return w
		},
	}
)

func init() {
	// init IP DB
	isCICD := os.Getenv("IS_CICD")
	if isCICD != "true" && !strings.HasSuffix(os.Args[0], ".test") {
		slog.Debug("Loading IP DB...")
		if ip, err := ip2location.OpenDB("./data/IPv4-DB11.BIN"); err != nil {
			slog.Error("Could not load IP DB file", "err", err)
			os.Exit(1)
		} else {
			ipDB = ip
		}
	} else {
		slog.Warn("Not loading IP DB")
	}

	// init Location
	if location, err := time.LoadLocation("America/Chicago"); err != nil {
		slog.Error("Could not load Chicago location", "err", err)
		os.Exit(1)
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
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusUnauthorized)
			if encodingErr := json.NewEncoder(res).Encode(err); encodingErr != nil {
				slog.Error("Could not encode API key error response", "err", encodingErr, "path", req.URL.Path)
			}
		} else {
			next.ServeHTTP(res, req)
		}
	})
}

// sets common headers for response
func commonResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("Cache-Control", "max-age=300")

		// gzip
		if acceptsGzip(req) {
			zip := gzipPool.Get().(*gzip.Writer)
			zip.Reset(res)
			defer gzipPool.Put(zip)
			defer zip.Close()

			res.Header().Set("Content-Encoding", "gzip")
			res.Header().Del("Content-Length")
			next.ServeHTTP(gzipResponseWriter{Writer: zip, ResponseWriter: res}, req)
		} else {
			next.ServeHTTP(res, req)
		}
	})
}

func acceptsGzip(req *http.Request) bool {
	for val := range strings.SplitSeq(req.Header.Get("Accept-Encoding"), ",") {
		if strings.TrimSpace(strings.Split(val, ";")[0]) == "gzip" {
			return true
		}
	}
	return false
}

// Configures routes and their middle wares
// This method should be called before the environment is set up as the API Key will be set according to the value found in environment
func RunHttpServer() {
	serverAPIKey = cUtil.EnvMap["API_KEY"] // configure API Key
	router := chi.NewRouter()

	// common middleware
	router.Use(commonResponseMiddleware)

	router.Route(v1Context, func(r chi.Router) {
		// configure non-admin routes
		r.Group(func(r chi.Router) {
			r.Get("/status", getAPIStatusHandler)
			r.Post("/card-details", getBatchCardInfo)
			r.Get("/card-of-the-day", getCardOfTheDay)

			// suggestions
			r.Get(`/card/{cardID:\d{8}}`, getCardSuggestionsHandler)
			r.Post("/card", getBatchSuggestionsHandler)

			// support
			r.Get(`/card/support/{cardID:\d{8}}`, getCardSupportHandler)
			r.Post("/card/support", getBatchSupportHandler)

			// similar resources
			r.Get(`/card/{cardID:\d{8}}/similar`, getSimilarCardsHandler)

			r.Get(`/product/{productID:[0-9A-Z]{3,4}}`, getProductSuggestionsHandler)
			r.Get("/archetype/{archetypeName}", getArchetypeSupportHandler)
			r.Get(`/trending/{resource:(?i)card|product}`, trending)
		})

		// admin routes
		r.Group(func(r chi.Router) {
			r.Use(verifyAPIKeyMiddleware)
			r.Post("/traffic-analysis", submitNewTrafficDataHandler)
		})
	})

	router.Route(v2Context, func(r chi.Router) {
		// configure non-admin routes
		r.Group(func(r chi.Router) {
			r.Get("/archetype/{archetypeName}", getArchetypeSupportV2Handler)
		})
	})

	// Cors
	corsOpts := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"https://dev.thesupremekingscastle.com",
			"https://thesupremekingscastle.com",
			"https://www.thesupremekingscastle.com",
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodOptions,
		},

		AllowedHeaders: []string{
			"*",
		},
	})

	cUtil.CombineCerts("certs")
	serveTLS(router, corsOpts)
}

// Configures and starts an HTTPS server with TLS encryption.
// It combines the TLS certificate and CA bundle, and utilizes the private key.
// Finally, it applies CORS middleware.
func serveTLS(router *chi.Mux, corsOpts *cors.Cors) {
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
		NextProtos: []string{"h2"},
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
	}

	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", apiPort),
		Handler:   corsOpts.Handler(router),
		TLSConfig: tlsCfg,

		ReadTimeout:       6 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      4 * time.Second,
		IdleTimeout:       15 * time.Second,

		MaxHeaderBytes: 32 << 10,
	}

	if err := http2.ConfigureServer(server, &http2.Server{
		MaxConcurrentStreams:         100,
		MaxHandlers:                  25,
		IdleTimeout:                  15 * time.Second,
		WriteByteTimeout:             4 * time.Second,
		MaxUploadBufferPerConnection: 20 << 10,
		MaxUploadBufferPerStream:     4 << 10,
	}); err != nil {
		slog.Error("Failed to configure HTTP/2", "err", err)
		os.Exit(1)
	}

	slog.Info("API starting", "port", apiPort)

	if err := server.ListenAndServeTLS("certs/concatenated.crt", "certs/private.key"); err != nil {
		slog.Error("There was an error starting api server", "err", err)
		os.Exit(1)
	}
}
