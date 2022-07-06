package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	skcDBConn                             *sql.DB
	client                                *mongo.Client
	skcSuggestionEngineDeckListCollection *mongo.Collection
)

// Connect to SKC database.
func EstablishSKCDBConn() {
	uri := "%s:%s@tcp(%s)/%s"
	dataSourceName := fmt.Sprintf(uri, env.EnvMap["SKC_DB_USER"], env.EnvMap["SKC_DB_PWD"], env.EnvMap["SKC_DB_URI"], env.EnvMap["SKC_DB_NAME"])

	var err error
	if skcDBConn, err = sql.Open("mysql", dataSourceName); err != nil {
		log.Fatalln("Error occurred while trying to establish DB connection: ", err)
	}
}

func EstablishSKCSuggestionEngineDBConn() {
	certificateKeyFilePath := "./certs/skc-suggestion-engine-db.pem"
	uri := "mongodb+srv://skc-suggestion-engine-e.rfait.mongodb.net/?tlsCertificateKeyFile=%s"
	uri = fmt.Sprintf(uri, certificateKeyFilePath)

	credential := options.Credential{
		AuthMechanism: "MONGODB-X509",
	}

	var err error

	if client, err = mongo.NewClient(options.Client().ApplyURI(uri).SetAuth(credential).SetMinPoolSize(40).SetMaxPoolSize(75).SetConnectTimeout(1 * time.Second)); err != nil {
		log.Fatalln("Error creating new mongodb client for skc-suggestion-engine", err)
	}

	if err = client.Connect(context.TODO()); err != nil {
		log.Fatal("Error connecting to skc-suggestion-engine DB", err)
	}

	skcSuggestionEngineDeckListCollection = client.Database("suggestions").Collection("deckList")
}
