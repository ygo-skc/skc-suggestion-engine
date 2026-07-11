package db

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

const (
	minPoolSize = 15
	maxPoolSize = 50

	certificateKeyFilePath = "./certs/skc-suggestion-engine-db.pem"
	connectTimeout         = 2 * time.Second
	serverSelectionTimeout = 5 * time.Second
)

var (
	skcSuggestionDB           *mongo.Database
	blackListCollection       *mongo.Collection
	trafficAnalysisCollection *mongo.Collection
	cardOfTheDayCollection    *mongo.Collection
	cardEmbeddingCollection   *mongo.Collection
)

func EstablishSKCSuggestionEngineDBConn() {
	uri := fmt.Sprintf("%s/?tlsCertificateKeyFile=%s", cUtil.EnvMap["DB_HOST"], certificateKeyFilePath)

	credential := options.Credential{
		AuthMechanism: "MONGODB-X509",
	}

	client, err := mongo.Connect(options.Client().
		ApplyURI(uri).
		SetAuth(credential).
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)).
		SetMinPoolSize(minPoolSize).
		SetMaxPoolSize(maxPoolSize).
		SetMaxConnIdleTime(10 * time.Minute).
		SetConnectTimeout(connectTimeout).
		SetServerSelectionTimeout(serverSelectionTimeout).
		SetTimeout(2 * time.Second).
		SetRetryWrites(true).
		SetRetryReads(true).
		SetReadConcern(readconcern.Available()).  // prefer eventually consistent reads
		SetWriteConcern(writeconcern.Majority()). // writes to most replicas before acknowledging the write is complete
		SetAppName("SKC Suggestion Engine"))
	if err != nil {
		log.Fatalln("Error creating new mongodb client for skc-suggestion-engine DB", err)
	}

	// ping mongo on startup
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalln("Error connecting to skc-suggestion-engine DB", err)
	}

	skcSuggestionDB = client.Database("suggestionDB")

	// init collections
	blackListCollection = skcSuggestionDB.Collection("blackList")
	trafficAnalysisCollection = skcSuggestionDB.Collection("trafficAnalysis")
	cardOfTheDayCollection = skcSuggestionDB.Collection("cardOfTheDay")
	cardEmbeddingCollection = skcSuggestionDB.Collection("cardEmbedding")

	slog.Info("Connected to suggestion engine DB")
}
