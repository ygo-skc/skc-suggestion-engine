package db

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

const (
	minPoolSize = 15
	maxPoolSize = 30

	certificateKeyFilePath = "./certs/skc-suggestion-engine-db.pem"
	connectTimeout         = 2 * time.Second
	serverSelectionTimeout = 5 * time.Second
)

var (
	skcSuggestionDB           *mongo.Database
	blackListCollection       *mongo.Collection
	trafficAnalysisCollection *mongo.Collection
	cardOfTheDayCollection    *mongo.Collection
	archetypeCollection       *mongo.Collection

	vectorSearchDB          *mongo.Database
	cardEmbeddingCollection *mongo.Collection

	// shared across both connections; per-connection ReadConcern is layered on top via mongo.Connect's option merging
	baseOpts = options.Client().
			SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)).
			SetMinPoolSize(minPoolSize).
			SetMaxPoolSize(maxPoolSize).
			SetMaxConnIdleTime(10 * time.Minute).
			SetConnectTimeout(connectTimeout).
			SetServerSelectionTimeout(serverSelectionTimeout).
			SetTimeout(2 * time.Second).
			SetRetryWrites(true).
			SetRetryReads(true).
			SetWriteConcern(writeconcern.Majority()). // writes to most replicas before acknowledging the write is complete
			SetAppName("SKC Suggestion Engine")
)

func EstablishSKCSuggestionEngineDBConn() {
	uri := fmt.Sprintf("%s/?tlsCertificateKeyFile=%s", cUtil.EnvMap["DB_HOST"], certificateKeyFilePath)
	credential := options.Credential{
		AuthMechanism: "MONGODB-X509",
	}

	// general purpose connection - prefer eventually consistent reads
	generalClient := connect(uri, credential, readconcern.Available())
	skcSuggestionDB = generalClient.Database("suggestionDB")
	blackListCollection = skcSuggestionDB.Collection("blackList")
	trafficAnalysisCollection = skcSuggestionDB.Collection("trafficAnalysis")
	cardOfTheDayCollection = skcSuggestionDB.Collection("cardOfTheDay")
	archetypeCollection = skcSuggestionDB.Collection("archetype")

	// vector search connection - $vectorSearch aggregation stage requires ReadConcern local
	vectorSearchClient := connect(uri, credential, readconcern.Local())
	vectorSearchDB = vectorSearchClient.Database("suggestionDB")
	cardEmbeddingCollection = vectorSearchDB.Collection("cardEmbedding")

	if err := createIndexes(); err != nil {
		slog.Error("Error creating indexes for skc-deck-api-db", "err", err)
		os.Exit(1)
	}

	slog.Info("Connected to suggestion engine DB")
}

func connect(uri string, credential options.Credential, rc *readconcern.ReadConcern) *mongo.Client {
	opts := options.Client().
		ApplyURI(uri).
		SetAuth(credential).
		SetReadConcern(rc)

	client, err := mongo.Connect(baseOpts, opts)
	if err != nil {
		log.Fatalln("Error creating new mongodb client for skc-suggestion-engine DB", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalln("Error connecting to skc-suggestion-engine DB", err)
	}

	return client
}

func createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexesByCollection := map[*mongo.Collection][]mongo.IndexModel{
		blackListCollection: {
			{
				Keys:    bson.D{{Key: "type", Value: 1}, {Key: "phrase", Value: 1}},
				Options: options.Index().SetName("blacklist_type_and_phrase").SetUnique(true),
			},
		},
		archetypeCollection: {
			{
				Keys:    bson.D{{Key: "archetype", Value: 1}},
				Options: options.Index().SetName("archetype_name").SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "inheritMembers", Value: 1}},
				Options: options.Index().SetName("archetype_inherit_members"),
			},
			{
				Keys:    bson.D{{Key: "qualifiedMembers", Value: 1}},
				Options: options.Index().SetName("archetype_qualified_members"),
			},
		},
	}

	for collection, indexes := range indexesByCollection {
		_, err := collection.Indexes().CreateMany(ctx, indexes)
		if err != nil {
			return fmt.Errorf("error creating indexes for collection %s: %w", collection.Name(), err)
		}
	}
	return nil
}
