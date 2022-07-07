package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/util"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	skcDBConn                             *sql.DB
	client                                *mongo.Client
	skcSuggestionEngineDeckListCollection *mongo.Collection
)

const (
	minPoolSize = 20
	maxPoolSize = 40
)

// Connect to SKC database.
func EstablishSKCDBConn() {
	uri := "%s:%s@tcp(%s)/%s"
	dataSourceName := fmt.Sprintf(uri, util.EnvMap["SKC_DB_USER"], util.EnvMap["SKC_DB_PWD"], util.EnvMap["SKC_DB_URI"], util.EnvMap["SKC_DB_NAME"])

	var err error
	if skcDBConn, err = sql.Open("mysql", dataSourceName); err != nil {
		log.Fatalln("Error occurred while trying to establish DB connection: ", err)
	}

	skcDBConn.SetMaxIdleConns(minPoolSize)
	skcDBConn.SetMaxOpenConns(maxPoolSize)
}

func EstablishSKCSuggestionEngineDBConn() {
	certificateKeyFilePath := "./certs/skc-suggestion-engine-db.pem"
	uri := "mongodb+srv://skc-suggestion-engine-e.rfait.mongodb.net/?tlsCertificateKeyFile=%s"
	uri = fmt.Sprintf(uri, certificateKeyFilePath)

	credential := options.Credential{
		AuthMechanism: "MONGODB-X509",
	}

	var err error

	if client, err = mongo.NewClient(options.Client().ApplyURI(uri).SetAuth(credential).SetMinPoolSize(minPoolSize).SetMaxPoolSize(maxPoolSize).SetConnectTimeout(1 * time.Second)); err != nil {
		log.Fatalln("Error creating new mongodb client for skc-suggestion-engine", err)
	}

	if err = client.Connect(context.TODO()); err != nil {
		log.Fatal("Error connecting to skc-suggestion-engine DB", err)
	}

	skcSuggestionEngineDeckListCollection = client.Database("suggestions").Collection("deckList")
}
