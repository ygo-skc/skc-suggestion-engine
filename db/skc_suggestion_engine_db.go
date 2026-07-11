package db

import (
	"context"
	"fmt"
	"net/http"
	"time"

	cModel "github.com/ygo-skc/skc-go/common/v2/model"
	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	intervalFormat = "2006-01-02"
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
	GetSimilarCards(context.Context, cModel.YGOCard) ([]model.VectorSearchResult, *cModel.APIError)
}

// impl
type SKCSuggestionEngineDAOImplementation struct{}

// Retrieves the version number of the SKC Suggestion DB or throws an error if an exception occurs.
func (impl SKCSuggestionEngineDAOImplementation) GetSKCSuggestionDBVersion(ctx context.Context) (string, error) {
	var commandResult bson.M
	command := bson.D{{Key: "serverStatus", Value: 1}}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := skcSuggestionDB.RunCommand(ctx, command).Decode(&commandResult); err != nil {
		cUtil.RetrieveLogger(ctx).Error("Error getting SKC Suggestion DB version", "err", err)
		return "", err
	} else {
		return fmt.Sprintf("%v", commandResult["version"]), nil
	}
}

// Will update the database with a new traffic record.
func (impl SKCSuggestionEngineDAOImplementation) InsertTrafficData(ctx context.Context, ta model.TrafficAnalysis) *cModel.APIError {
	logger := cUtil.RetrieveLogger(ctx)
	logger.Info("Inserting traffic data", "resource", ta.ResourceUtilized, "system", ta.Source)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if res, err := trafficAnalysisCollection.InsertOne(ctx, ta); err != nil {
		logger.Error("Error inserting traffic data into DB", "err", err)
		return &cModel.APIError{Message: "Error occurred while attempting to insert new traffic data.", StatusCode: http.StatusInternalServerError}
	} else {
		logger.Info("Successfully inserted traffic data into DB", "id", res.InsertedID)
		return nil
	}
}

func (impl SKCSuggestionEngineDAOImplementation) GetTrafficData(
	ctx context.Context, resourceName model.ResourceName, from time.Time, to time.Time) ([]model.TrafficResourceUtilizationMetric, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{
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
		{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: "$resourceUtilized.value"},
					{Key: "occurrences", Value: bson.D{{Key: "$sum", Value: 1}}},
				},
			},
		},
		{
			{Key: "$sort", Value: bson.D{
				{Key: "occurrences", Value: -1},
				{Key: "_id", Value: -1},
			}}},
		{{Key: "$limit", Value: 10}},
	}

	if cursor, err := trafficAnalysisCollection.Aggregate(ctx, pipeline); err != nil {
		logger.Error("Error retrieving traffic data",
			"resource", resourceName, "from", from.Format(intervalFormat), "to", to.Format(intervalFormat), "err", err)
		return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Could not get traffic data."}
	} else {
		td := []model.TrafficResourceUtilizationMetric{}
		if err := cursor.All(ctx, &td); err != nil {
			logger.Error("Error retrieving traffic data",
				"resource", resourceName, "from", from.Format(intervalFormat), "to", to.Format(intervalFormat), "err", err)
			return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Could not get traffic data."}
		}

		return td, nil
	}
}

func (impl SKCSuggestionEngineDAOImplementation) IsBlackListed(ctx context.Context, blackListType string, blackListPhrase string) (bool, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	query := bson.M{"type": blackListType, "phrase": blackListPhrase}

	if count, err := blackListCollection.CountDocuments(ctx, query); err != nil {
		message := fmt.Sprintf("Black list query failed using type %s and phrase %s.", blackListType, blackListPhrase)
		return false, &cModel.APIError{Message: message, StatusCode: http.StatusInternalServerError}
	} else if count > 0 {
		logger.Info("Phrase is blacklisted", "phrase", blackListPhrase)
		return true, nil
	} else {
		return false, nil
	}
}

func (impl SKCSuggestionEngineDAOImplementation) GetCardOfTheDay(ctx context.Context, date string, version int) (*string, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
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
		logger.Error("Error retrieving card of the day", "date", date, "err", err)
		return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Could not get card of the day."}
	}

	return &cotd.CardID, nil
}

func (impl SKCSuggestionEngineDAOImplementation) GetHistoricalCardOfTheDayData(ctx context.Context, version int) ([]string, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	query := bson.M{"version": version}
	opts := options.Find().SetProjection( // select only these fields from collection
		bson.D{
			{Key: "cardID", Value: 1},
		},
	)

	if cursor, err := cardOfTheDayCollection.Find(ctx, query, opts); err != nil {
		logger.Error("Error retrieving card of the day history", "version", version, "err", err)
	} else {
		historicalCOTD := make([]string, 0, 100)
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var cotd model.CardOfTheDay
			if err := cursor.Decode(&cotd); err != nil {
				logger.Error("Error transforming DB data to COTD struct", "version", version, "err", err)
			}
			historicalCOTD = append(historicalCOTD, cotd.CardID)
		}
		return historicalCOTD, nil
	}
	return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error retrieving card of the day history"}
}

func (impl SKCSuggestionEngineDAOImplementation) InsertCardOfTheDay(ctx context.Context, cotd model.CardOfTheDay) *cModel.APIError {
	logger := cUtil.RetrieveLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	logger.Info("Inserting new COTD", "id", cotd.CardID, "version", cotd.Version)

	if _, err := cardOfTheDayCollection.InsertOne(ctx, cotd); err != nil {
		logger.Error("Could not insert card of the day", "err", err)
		return &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error saving card of the day."}
	}

	logger.Info("Successfully inserted new card of the day.")
	return nil
}

func (impl SKCSuggestionEngineDAOImplementation) GetSimilarCards(ctx context.Context,
	subject cModel.YGOCard) ([]model.VectorSearchResult, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	logger.Info("Performing vector search on card")

	desiredResults := 20
	subjectEffect := subject.GetEffect()

	pipeline := mongo.Pipeline{
		{
			{
				Key: "$vectorSearch", Value: bson.D{
					{Key: "index", Value: "text_embedding"},
					{Key: "path", Value: "text"},
					{Key: "query", Value: bson.D{
						{Key: "text", Value: subjectEffect},
					}},
					{Key: "numCandidates", Value: 100},
					{Key: "limit", Value: 30},
				},
			},
		},
		{
			{
				Key: "$rerank", Value: bson.D{
					{Key: "model", Value: "rerank-2.5"},
					{Key: "query", Value: bson.D{
						{Key: "text", Value: subjectEffect}, // TODO: update rerank query
					}},
					{Key: "path", Value: "text"},
					{Key: "numDocsToRerank", Value: 30},
				},
			},
		},
		{
			{Key: "$limit", Value: desiredResults},
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "id", Value: 1},
			}},
		},
	}

	cursor, err := cardEmbeddingCollection.Aggregate(ctx, pipeline)
	if err != nil {
		logger.Error("Error retrieving similar card", "err", err)
		return make([]model.VectorSearchResult, 0, 0), &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error retrieving similar card"}
	}

	defer cursor.Close(ctx)

	results := make([]model.VectorSearchResult, 0, desiredResults)
	for cursor.Next(ctx) {
		var r model.VectorSearchResult
		if err := cursor.Decode(&r); err != nil {
			logger.Error("Error transforming DB data to Vector Search struct", "err", err)
		}
		results = append(results, r)
	}
	logger.Info("results", "res", results)

	return results, nil
}
