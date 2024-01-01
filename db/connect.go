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

const (
	minPoolSize = 20
	maxPoolSize = 40
)

// Connect to SKC database.
func EstablishDBConn() {
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

	if client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri).SetAuth(credential).SetMinPoolSize(minPoolSize).SetMaxPoolSize(maxPoolSize).SetMaxConnIdleTime(10*time.Minute).
		SetAppName("SKC Suggestion Engine")); err != nil {
		log.Fatalln("Error creating new mongodb client for skc-suggestion-engine DB", err)
	} else {
		skcSuggestionDB = client.Database("suggestionDB")
	}

	// init collections
	blackListCollection = skcSuggestionDB.Collection("blackList")
	deckListCollection = skcSuggestionDB.Collection("deckLists")
	trafficAnalysisCollection = skcSuggestionDB.Collection("trafficAnalysis")
	cardOfTheDayCollection = skcSuggestionDB.Collection("cardOfTheDay")
}
