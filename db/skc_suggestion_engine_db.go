package db

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	skcSuggestionDB           *mongo.Database
	blackListCollection       *mongo.Collection
	trafficAnalysisCollection *mongo.Collection
	cardOfTheDayCollection    *mongo.Collection
)

// interface
type SKCSuggestionEngineDAO interface {
	GetSKCSuggestionDBVersion(ctx context.Context) (string, error)

	InsertTrafficData(ctx context.Context, ta model.TrafficAnalysis) *model.APIError
	GetTrafficData(ctx context.Context, resourceName model.ResourceName, from time.Time, to time.Time) ([]model.TrafficResourceUtilizationMetric, *model.APIError)

	IsBlackListed(ctx context.Context, blackListType string, blackListPhrase string) (bool, *model.APIError)

	GetCardOfTheDay(ctx context.Context, date string) (*string, *model.APIError)
	InsertCardOfTheDay(ctx context.Context, cotd model.CardOfTheDay) *model.APIError
}

// impl
type SKCSuggestionEngineDAOImplementation struct{}

// Retrieves the version number of the SKC Suggestion DB or throws an error if an exception occurs.
func (dbInterface SKCSuggestionEngineDAOImplementation) GetSKCSuggestionDBVersion(ctx context.Context) (string, error) {
	var commandResult bson.M
	command := bson.D{{Key: "serverStatus", Value: 1}}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := skcSuggestionDB.RunCommand(ctx, command).Decode(&commandResult); err != nil {
		util.LoggerFromContext(ctx).Error(fmt.Sprintf("Error getting SKC Suggestion DB version %v", err))
		return "", err
	} else {
		return fmt.Sprintf("%v", commandResult["version"]), nil
	}
}

// Will update the database with a new traffic record.
func (dbInterface SKCSuggestionEngineDAOImplementation) InsertTrafficData(ctx context.Context, ta model.TrafficAnalysis) *model.APIError {
	logger := util.LoggerFromContext(ctx)
	logger.Info(fmt.Sprintf("Inserting traffic data for resource %+v and system %+v.", ta.ResourceUtilized, ta.Source))

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if res, err := trafficAnalysisCollection.InsertOne(ctx, ta); err != nil {
		logger.Error(fmt.Sprintf("Error inserting traffic data into DB: %v", err))
		return &model.APIError{Message: "Error occurred while attempting to insert new traffic data.", StatusCode: http.StatusInternalServerError}
	} else {
		logger.Info(fmt.Sprintf("Successfully inserted traffic data into DB, ID: %v", res.InsertedID))
		return nil
	}
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetTrafficData(
	ctx context.Context, resourceName model.ResourceName, from time.Time, to time.Time) ([]model.TrafficResourceUtilizationMetric, *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
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
		logger.Error(fmt.Sprintf("Error retrieving traffic data for resource w/ name %s. Err: %v", resourceName, err))
		return nil, &model.APIError{Message: "Could not get traffic data."}
	} else {
		td := []model.TrafficResourceUtilizationMetric{}
		if err := cursor.All(ctx, &td); err != nil {
			logger.Error(fmt.Sprintf("Error retrieving traffic data for resource w/ name %s. Err: %v", resourceName, err))
			return nil, &model.APIError{Message: "Could not get traffic data."}
		}

		return td, nil
	}
}

func (dbInterface SKCSuggestionEngineDAOImplementation) IsBlackListed(ctx context.Context, blackListType string, blackListPhrase string) (bool, *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	query := bson.M{"type": blackListType, "phrase": blackListPhrase}

	if count, err := blackListCollection.CountDocuments(ctx, query); err != nil {
		message := fmt.Sprintf("Black list query failed using type %s and phrase %s.", blackListType, blackListPhrase)
		return false, &model.APIError{Message: message, StatusCode: http.StatusInternalServerError}
	} else if count > 0 {
		logger.Info(fmt.Sprintf("%s is blacklisted", blackListPhrase))
		return true, nil
	} else {
		return false, nil
	}
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetCardOfTheDay(ctx context.Context, date string) (*string, *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
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
		logger.Error(fmt.Sprintf("Error retrieving card of the day for given date: %s. Err: %s", date, err))
		return nil, &model.APIError{Message: "Could not get card of the day."}
	}

	return &cotd.CardID, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) InsertCardOfTheDay(ctx context.Context, cotd model.CardOfTheDay) *model.APIError {
	logger := util.LoggerFromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	logger.Info(fmt.Sprintf("Inserting new card of the day for date %s. Card being saved is %s. Using version %d.", cotd.Date, cotd.CardID, cotd.Version))

	if _, err := cardOfTheDayCollection.InsertOne(ctx, cotd); err != nil {
		logger.Error(fmt.Sprintf("Could not insert card of the day, err %s", err))
		return &model.APIError{StatusCode: http.StatusInternalServerError, Message: "Error saving card of the day."}
	}

	logger.Info("Successfully inserted new card of the day.")
	return nil
}
