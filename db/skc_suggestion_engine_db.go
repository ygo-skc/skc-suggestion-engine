package db

import (
	"context"
	"fmt"
	"net/http"
	"time"

	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	skcSuggestionDB           *mongo.Database
	blackListCollection       *mongo.Collection
	trafficAnalysisCollection *mongo.Collection
	cardOfTheDayCollection    *mongo.Collection
)

// interface
type SKCSuggestionEngineDAO interface {
	GetSKCSuggestionDBVersion(context.Context) (string, error)

	InsertTrafficData(context.Context, model.TrafficAnalysis) *cModel.APIError
	GetTrafficData(context.Context, model.ResourceName, time.Time, time.Time) ([]model.TrafficResourceUtilizationMetric, *cModel.APIError)

	IsBlackListed(context.Context, string, string) (bool, *cModel.APIError)

	GetCardOfTheDay(context.Context, string, int) (*string, *cModel.APIError)
	GetHistoricalCardOfTheDayData(context.Context, int) ([]string, *cModel.APIError)
	InsertCardOfTheDay(context.Context, model.CardOfTheDay) *cModel.APIError
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
		cUtil.LoggerFromContext(ctx).Error(fmt.Sprintf("Error getting SKC Suggestion DB version %v", err))
		return "", err
	} else {
		return fmt.Sprintf("%v", commandResult["version"]), nil
	}
}

// Will update the database with a new traffic record.
func (dbInterface SKCSuggestionEngineDAOImplementation) InsertTrafficData(ctx context.Context, ta model.TrafficAnalysis) *cModel.APIError {
	logger := cUtil.LoggerFromContext(ctx)
	logger.Info(fmt.Sprintf("Inserting traffic data for resource %+v and system %+v.", ta.ResourceUtilized, ta.Source))

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if res, err := trafficAnalysisCollection.InsertOne(ctx, ta); err != nil {
		logger.Error(fmt.Sprintf("Error inserting traffic data into DB: %v", err))
		return &cModel.APIError{Message: "Error occurred while attempting to insert new traffic data.", StatusCode: http.StatusInternalServerError}
	} else {
		logger.Info(fmt.Sprintf("Successfully inserted traffic data into DB, ID: %v", res.InsertedID))
		return nil
	}
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetTrafficData(
	ctx context.Context, resourceName model.ResourceName, from time.Time, to time.Time) ([]model.TrafficResourceUtilizationMetric, *cModel.APIError) {
	logger := cUtil.LoggerFromContext(ctx)
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
		return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Could not get traffic data."}
	} else {
		td := []model.TrafficResourceUtilizationMetric{}
		if err := cursor.All(ctx, &td); err != nil {
			logger.Error(fmt.Sprintf("Error retrieving traffic data for resource w/ name %s. Err: %v", resourceName, err))
			return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Could not get traffic data."}
		}

		return td, nil
	}
}

func (dbInterface SKCSuggestionEngineDAOImplementation) IsBlackListed(ctx context.Context, blackListType string, blackListPhrase string) (bool, *cModel.APIError) {
	logger := cUtil.LoggerFromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	query := bson.M{"type": blackListType, "phrase": blackListPhrase}

	if count, err := blackListCollection.CountDocuments(ctx, query); err != nil {
		message := fmt.Sprintf("Black list query failed using type %s and phrase %s.", blackListType, blackListPhrase)
		return false, &cModel.APIError{Message: message, StatusCode: http.StatusInternalServerError}
	} else if count > 0 {
		logger.Info(fmt.Sprintf("%s is blacklisted", blackListPhrase))
		return true, nil
	} else {
		return false, nil
	}
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetCardOfTheDay(ctx context.Context, date string, version int) (*string, *cModel.APIError) {
	logger := cUtil.LoggerFromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	query := bson.M{"date": date, "version": version}
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
		return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Could not get card of the day."}
	}

	return &cotd.CardID, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetHistoricalCardOfTheDayData(ctx context.Context, version int) ([]string, *cModel.APIError) {
	logger := cUtil.LoggerFromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	query := bson.M{"version": version}
	opts := options.Find().SetProjection( // select only these fields from collection
		bson.D{
			{Key: "cardID", Value: 1},
		},
	)

	if cursor, err := cardOfTheDayCollection.Find(ctx, query, opts); err != nil {
		logger.Error(fmt.Sprintf("Error retrieving card of the day history. Version: %d. Err: %s", version, err))
	} else {
		historicalCOTD := make([]string, 0, 100)
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var cotd model.CardOfTheDay
			if err := cursor.Decode(&cotd); err != nil {
				logger.Error(fmt.Sprintf("Error transforming DB data to COTD struct. Version: %d. Err: %s", version, err))
			}
			historicalCOTD = append(historicalCOTD, cotd.CardID)
		}
		return historicalCOTD, nil
	}
	return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error retrieving card of the day history"}
}

func (dbInterface SKCSuggestionEngineDAOImplementation) InsertCardOfTheDay(ctx context.Context, cotd model.CardOfTheDay) *cModel.APIError {
	logger := cUtil.LoggerFromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	logger.Info(fmt.Sprintf("Inserting new COTD - ID: %s version %d", cotd.CardID, cotd.Version))

	if _, err := cardOfTheDayCollection.InsertOne(ctx, cotd); err != nil {
		logger.Error(fmt.Sprintf("Could not insert card of the day, err %s", err))
		return &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error saving card of the day."}
	}

	logger.Info("Successfully inserted new card of the day.")
	return nil
}
