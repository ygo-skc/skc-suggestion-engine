package db

import (
	"fmt"
	"log"
	"time"

	cUtil "github.com/ygo-skc/skc-go/common/util"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

const (
	minPoolSize = 5
	maxPoolSize = 50
)

func EstablishSKCSuggestionEngineDBConn() {
	certificateKeyFilePath := "./certs/skc-suggestion-engine-db.pem"
	uri := fmt.Sprintf("%s/?tlsCertificateKeyFile=%s", cUtil.EnvMap["DB_HOST"], certificateKeyFilePath)

	credential := options.Credential{
		AuthMechanism: "MONGODB-X509",
	}

	if client, err := mongo.Connect(options.Client().
		ApplyURI(uri).
		SetAuth(credential).
		SetMinPoolSize(minPoolSize).
		SetMaxPoolSize(maxPoolSize).
		SetMaxConnIdleTime(20 * time.Minute).
		SetTimeout(2 * time.Second).
		SetReadConcern(readconcern.Majority()).   // prefer strongly consistent reeds
		SetWriteConcern(writeconcern.Majority()). // writes to most replicas before acknowledging the write is complete
		SetCompressors([]string{"zlib"}).
		SetAppName("SKC Suggestion Engine")); err != nil {
		log.Fatalln("Error creating new mongodb client for skc-suggestion-engine DB", err)
	} else {
		skcSuggestionDB = client.Database("suggestionDB")
	}

	// init collections
	blackListCollection = skcSuggestionDB.Collection("blackList")
	trafficAnalysisCollection = skcSuggestionDB.Collection("trafficAnalysis")
	cardOfTheDayCollection = skcSuggestionDB.Collection("cardOfTheDay")
}
