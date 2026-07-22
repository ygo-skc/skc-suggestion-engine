package db

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	cModel "github.com/ygo-skc/skc-go/common/v3/model"
	cUtil "github.com/ygo-skc/skc-go/common/v3/util"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	intervalFormat = "2006-01-02"

	maxBlackListPhraseLength = 40
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

	GetArchetypeMembers(context.Context, string) ([]string, []string, []string, *cModel.APIError)
	GetRelevantArchetypes(context.Context, cModel.CardIDs) ([]string, *cModel.APIError)

	VectorSearchOnCardEmbedding(context.Context, cModel.YGOCard, []float32) ([]model.VectorSearchResult, *cModel.APIError)
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

	blackListPhrase = sanitizeQueryInput(blackListPhrase)
	if blackListPhrase == "" || len(blackListPhrase) > maxBlackListPhraseLength {
		logger.Error("Rejecting black list check for invalid phrase", "type", blackListType, "phrase_length", len(blackListPhrase))
		return false, &cModel.APIError{Message: "Internal server error", StatusCode: http.StatusInternalServerError}
	}

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

// sanitizeQueryInput trims surrounding whitespace and strips control characters from user-supplied input before it is used to build a query.
func sanitizeQueryInput(s string) string {
	s = strings.TrimSpace(s)
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
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
		if errors.Is(err, mongo.ErrNoDocuments) { // no card of the day present in db for specified date
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

	cursor, err := cardOfTheDayCollection.Find(ctx, query, opts)
	if err != nil {
		logger.Error("Error retrieving card of the day history", "version", version, "err", err)
		return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error retrieving card of the day history"}
	}
	defer cursor.Close(ctx)

	historicalCOTD := make([]string, 0, 100)
	for cursor.Next(ctx) {
		var cotd model.CardOfTheDay
		if err := cursor.Decode(&cotd); err != nil {
			logger.Error("Error transforming DB data to COTD struct", "version", version, "err", err)
		}
		historicalCOTD = append(historicalCOTD, cotd.CardID)
	}

	// check if there was an error using cursor
	if err := cursor.Err(); err != nil {
		logger.Error("Error iterating card of the day history", "version", version, "err", err)
		return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error retrieving card of the day history"}
	}
	return historicalCOTD, nil
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

func (impl SKCSuggestionEngineDAOImplementation) GetArchetypeMembers(ctx context.Context, archetype string) ([]string, []string, []string, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	logger.Info("Fetching archetype members")

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	query := bson.M{"archetype": archetype}

	type archetypeMembers struct {
		Archetype        string   `bson:"archetype"`
		InheritMembers   []string `bson:"inheritMembers"`
		QualifiedMembers []string `bson:"qualifiedMembers"`
		ExcludedMembers  []string `bson:"excludedMembers"`
	}

	var members archetypeMembers
	if err := archetypeCollection.FindOne(ctx, query).Decode(&members); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			logger.Error("Could not find archetype in DB", "err", err)
			return nil, nil, nil, &cModel.APIError{StatusCode: http.StatusNotFound, Message: "Archetype does not exist"}
		}
		logger.Error("Error retrieving archetype data", "err", err)
		return nil, nil, nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Could not get archetype data"}
	}

	return members.InheritMembers, members.QualifiedMembers, members.ExcludedMembers, nil
}

func (impl SKCSuggestionEngineDAOImplementation) GetRelevantArchetypes(ctx context.Context, subjects cModel.CardIDs) ([]string, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	query := bson.M{
		"$or": bson.A{
			bson.M{"inheritMembers": bson.M{"$in": subjects}},
			bson.M{"qualifiedMembers": bson.M{"$in": subjects}},
		},
	}

	opts := options.Find().SetProjection(
		bson.D{
			{Key: "_id", Value: 0},
			{Key: "archetype", Value: 1},
		},
	)

	cursor, err := archetypeCollection.Find(ctx, query, opts)
	if err != nil {
		logger.Error("Error retrieving relevant archetypes from DB", "err", err)
		return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error retrieving archetype data"}
	}
	defer cursor.Close(ctx)

	type relevantArchetypes struct {
		Archetype string `bson:"archetype"`
	}

	var ra []relevantArchetypes
	if err := cursor.All(ctx, &ra); err != nil {
		logger.Error("Error retrieving relevant archetypes from DB", "err", err)
		return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error retrieving archetype data"}
	}

	if err := cursor.Err(); err != nil {
		logger.Error("Error iterating relevant archetype DB data", "err", err)
		return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error retrieving archetype data"}
	}

	f := make([]string, len(ra))
	for i := range ra {
		f[i] = ra[i].Archetype
	}
	return f, nil
}

func (impl SKCSuggestionEngineDAOImplementation) VectorSearchOnCardEmbedding(ctx context.Context,
	subject cModel.YGOCard, queryVector []float32) ([]model.VectorSearchResult, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	logger.Info("Performing vector search on card text")

	limit := 30

	pipeline := mongo.Pipeline{
		{
			{
				Key: "$vectorSearch", Value: bson.D{
					{Key: "index", Value: "text_embedding"},
					{Key: "path", Value: "textEmbedding"},
					{Key: "exact", Value: true}, // true = ENN search https://www.mongodb.com/docs/vector-search/query/aggregation-stages/vector-search-stage/?deployment-type=atlas&embedding=auto&interface=driver&language=go#enn-search
					{Key: "filter", Value: bson.D{
						{Key: "id", Value: bson.D{
							{Key: "$ne", Value: subject.GetID()},
						}},
					}},
					{Key: "queryVector", Value: queryVector},
					{Key: "limit", Value: limit * 3},
				},
			},
		},
		{
			{Key: "$addFields", Value: bson.D{
				{Key: "cosineSimilarity", Value: bson.D{
					{Key: "$meta", Value: "vectorSearchScore"},
				}},
				{Key: "sharedType", Value: bson.D{
					{Key: "$cond", Value: bson.A{
						bson.D{{Key: "$eq", Value: bson.A{"$type", subject.GetMonsterType()}}},
						1,
						0,
					}},
				}},
				{Key: "sharedAttribute", Value: bson.D{
					{Key: "$cond", Value: bson.A{
						bson.D{{Key: "$eq", Value: bson.A{"$attribute", subject.GetAttribute()}}},
						1,
						0,
					}},
				}},
				{Key: "sharedMonsterType", Value: bson.D{
					{Key: "$cond", Value: bson.A{
						bson.D{{Key: "$eq", Value: bson.A{"$monsterType", subject.GetMonsterType()}}},
						1,
						0,
					}},
				}},
			}},
		},
		{
			{Key: "$addFields", Value: bson.D{
				{Key: "finalScore", Value: bson.D{
					{Key: "$add", Value: bson.A{
						"$cosineSimilarity",
						bson.D{{Key: "$multiply", Value: bson.A{"$sharedType", 0.03}}},
						bson.D{{Key: "$multiply", Value: bson.A{"$sharedAttribute", 0.05}}},
						bson.D{{Key: "$multiply", Value: bson.A{"$sharedMonsterType", 0.08}}},
					}},
				}},
			}},
		},
		{
			{Key: "$sort", Value: bson.D{
				{Key: "finalScore", Value: -1},
			}},
		},
		{
			{Key: "$limit", Value: limit},
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "id", Value: 1},
				{Key: "text", Value: 1},
				{Key: "cosineSimilarity", Value: 1},
				{Key: "sharedAttribute", Value: 1},
				{Key: "sharedMonsterType", Value: 1},
				{Key: "finalScore", Value: 1},
			}},
		},
	}

	cursor, err := cardEmbeddingCollection.Aggregate(ctx, pipeline)
	if err != nil {
		logger.Error("Error while searching card embedding", "err", err)
		return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error retrieving similar card"}
	}

	defer cursor.Close(ctx)

	results := make([]model.VectorSearchResult, 0, limit)
	for cursor.Next(ctx) {
		var r model.VectorSearchResult
		if err := cursor.Decode(&r); err != nil {
			logger.Error("Error transforming DB data to Vector Search struct", "err", err)
		}
		results = append(results, r)
	}

	// check if there was an error using cursor
	if err := cursor.Err(); err != nil {
		logger.Error("There was an error parsing db results", "err", err)
		return nil, &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "Error retrieving similar card"}
	}

	return results, nil
}
