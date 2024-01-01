package db

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	skcSuggestionDB           *mongo.Database
	blackListCollection       *mongo.Collection
	deckListCollection        *mongo.Collection
	trafficAnalysisCollection *mongo.Collection
	cardOfTheDayCollection    *mongo.Collection
)

// interface
type SKCSuggestionEngineDAO interface {
	GetSKCSuggestionDBVersion() (string, error)

	InsertTrafficData(ta model.TrafficAnalysis) *model.APIError
	GetTrafficData(resourceName model.ResourceName, from time.Time, to time.Time) ([]model.TrafficResourceUtilizationMetric, *model.APIError)

	IsBlackListed(blackListType string, blackListPhrase string) (bool, *model.APIError)

	GetCardOfTheDay(date string) (*string, *model.APIError)
	InsertCardOfTheDay(cotd model.CardOfTheDay) *model.APIError
}

// impl
type SKCSuggestionEngineDAOImplementation struct{}

// Retrieves the version number of the SKC Suggestion DB or throws an error if an exception occurs.
func (dbInterface SKCSuggestionEngineDAOImplementation) GetSKCSuggestionDBVersion() (string, error) {
	var commandResult bson.M
	command := bson.D{{Key: "serverStatus", Value: 1}}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := skcSuggestionDB.RunCommand(ctx, command).Decode(&commandResult); err != nil {
		log.Println("Error getting SKC Suggestion DB version", err)
		return "", err
	} else {
		return fmt.Sprintf("%v", commandResult["version"]), nil
	}
}

// Will update the database with a new traffic record.
func (dbInterface SKCSuggestionEngineDAOImplementation) InsertTrafficData(ta model.TrafficAnalysis) *model.APIError {
	log.Printf("Inserting traffic data for resource %+v and system %+v.", ta.ResourceUtilized, ta.Source)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if res, err := trafficAnalysisCollection.InsertOne(ctx, ta); err != nil {
		log.Printf("Error inserting traffic data into DB: %v", err)
		return &model.APIError{Message: "Error occurred while attempting to insert new traffic data.", StatusCode: http.StatusInternalServerError}
	} else {
		log.Printf("Successfully inserted traffic data into DB, ID: %v", res.InsertedID)
		return nil
	}
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetTrafficData(
	resourceName model.ResourceName, from time.Time, to time.Time) ([]model.TrafficResourceUtilizationMetric, *model.APIError) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	query := bson.A{
		bson.D{
			{Key: "$match",
				Value: bson.D{
					{Key: "resourceUtilized.name", Value: resourceName},
					{Key: "timestamp",
						Value: bson.D{
							{Key: "$gte", Value: from},
							{Key: "$lte", Value: to},
						},
					},
				},
			},
		},
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: "$resourceUtilized.value"},
					{Key: "occurrences", Value: bson.D{{Key: "$sum", Value: 1}}},
				},
			},
		},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "occurrences", Value: -1}, {Key: "_id", Value: 1},
		}}},
		bson.D{{Key: "$limit", Value: 10}},
	}

	if cursor, err := trafficAnalysisCollection.Aggregate(ctx, query); err != nil {
		log.Printf("Error retrieving traffic data for resource w/ name %s. Err: %v", resourceName, err)
		return nil, &model.APIError{Message: "Could not get traffic data."}
	} else {
		td := []model.TrafficResourceUtilizationMetric{}
		if err := cursor.All(ctx, &td); err != nil {
			log.Printf("Error retrieving traffic data for resource w/ name %s. Err: %v", resourceName, err)
			return nil, &model.APIError{Message: "Could not get traffic data."}
		}

		return td, nil
	}
}

func (dbInterface SKCSuggestionEngineDAOImplementation) IsBlackListed(blackListType string, blackListPhrase string) (bool, *model.APIError) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	query := bson.M{"type": blackListType, "phrase": blackListPhrase}

	if count, err := blackListCollection.CountDocuments(ctx, query); err != nil {
		message := fmt.Sprintf("Black list query failed using type %s and phrase %s.", blackListType, blackListPhrase)
		return false, &model.APIError{Message: message, StatusCode: http.StatusInternalServerError}
	} else if count > 0 {
		log.Printf("%s is blacklisted", blackListPhrase)
		return true, nil
	} else {
		return false, nil
	}
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetCardOfTheDay(date string) (*string, *model.APIError) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	query := bson.M{"date": date, "version": 1}
	opts := options.FindOne().SetProjection( // select only these fields from collection
		bson.D{
			{Key: "cardID", Value: 1},
		},
	)

	var cotd model.CardOfTheDay
	if err := cardOfTheDayCollection.FindOne(ctx, query, opts).Decode(&cotd); err != nil {
		if err.Error() == "mongo: no documents in result" { // no card of the day present in db for specified date
			return nil, nil
		}
		log.Printf("Error retrieving card of the day for given date: %s. Err: %s", date, err)
		return nil, &model.APIError{Message: "Could not get card of the day."}
	}

	return &cotd.CardID, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) InsertCardOfTheDay(cotd model.CardOfTheDay) *model.APIError {
	log.Printf("Inserting new card of the day for date %s. Card being saved is %s. Using version %d.", cotd.Date, cotd.CardID, cotd.Version)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if _, err := cardOfTheDayCollection.InsertOne(ctx, cotd); err != nil {
		log.Printf("Could not insert card of the day, err %s", err)
		return &model.APIError{StatusCode: http.StatusInternalServerError, Message: "Error saving card of the day."}
	}

	log.Println("Successfully inserted new card of the day.")
	return nil
}
