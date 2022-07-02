package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/ygo-skc/skc-suggestion-engine/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	skcDBConn                             *sql.DB
	client                                *mongo.Client
	ctx                                   context.Context
	skcSuggestionEngineDeckListCollection *mongo.Collection
)

// Connect to SKC database.
func EstablishSKCDBConn() {
	dataSourceName := env.EnvMap["SKC_DB_USER"] + ":" + env.EnvMap["SKC_DB_PWD"] + "@tcp(" + env.EnvMap["SKC_DB_URI"] + ")/" + env.EnvMap["SKC_DB_NAME"]

	var err error
	if skcDBConn, err = sql.Open("mysql", dataSourceName); err != nil {
		log.Fatalln("Error occurred while trying to establish DB connection: ", err)
	}
}

func EstablishSKCSuggestionEngineDBConn() {
	uri := "mongodb+srv://skc-suggestion-engine-e.rfait.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&tlsCertificateKeyFile=./certs/skc-suggestion-engine-db.pem"
	var err error

	if client, err = mongo.NewClient(options.Client().ApplyURI(uri)); err != nil {
		log.Fatalln("Error creating new mongodb client for skc-suggestion-engine", err)
	}

	ctx = context.Background()

	if err = client.Connect(ctx); err != nil {
		log.Fatal("Error connecting to skc-suggestion-engine DB", err)
	}

	skcSuggestionEngineDeckListCollection = client.Database("skc-suggestions").Collection("deck-list")
}
