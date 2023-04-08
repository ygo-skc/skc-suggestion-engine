package db

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	GetDecksThatFeatureCards([]string) (*[]model.DeckList, *model.APIError)
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

func InsertDeckList(deckList model.DeckList) {
	deckList.CreatedAt = time.Now()
	deckList.UpdatedAt = deckList.CreatedAt

	log.Printf("Inserting deck with name %s with Main Deck size %d and Extra Deck size %d. List contents (in base64 and possibly reformatted) %s", deckList.Name, deckList.NumMainDeckCards, deckList.NumExtraDeckCards, deckList.ContentB64)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if res, err := deckListCollection.InsertOne(ctx, deckList); err != nil {
		log.Println("Error inserting new deck list into DB", err)
	} else {
		log.Println("Successfully inserted new deck list into DB, ID:", res.InsertedID)
	}
}

func GetDeckList(deckID string) (*model.DeckList, *model.APIError) {
	if objectId, err := primitive.ObjectIDFromHex(deckID); err != nil {
		log.Println("Invalid Object ID.")
		return nil, &model.APIError{Message: "Object ID used for deck list was not valid."}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		var dl model.DeckList
		if err := deckListCollection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&dl); err != nil {
			log.Printf("Error retrieving deck list w/ ID %s. Err: %v", deckID, err)
			return nil, &model.APIError{Message: "Requested deck list not found in DB."}
		} else {
			return &dl, nil
		}
	}
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetDecksThatFeatureCards(cardIDs []string) (*[]model.DeckList, *model.APIError) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// select only these fields from collection
	opts := options.Find().SetProjection(
		bson.D{
			{Key: "name", Value: 1}, {Key: "videoUrl", Value: 1}, {Key: "uniqueCards", Value: 1}, {Key: "deckMascots", Value: 1}, {Key: "numMainDeckCards", Value: 1},
			{Key: "numExtraDeckCards", Value: 1}, {Key: "tags", Value: 1}, {Key: "createdAt", Value: 1}, {Key: "updatedAt", Value: 1},
		},
	)

	if cursor, err := deckListCollection.Find(ctx, bson.M{"uniqueCards": bson.M{"$in": cardIDs}}, opts); err != nil {
		log.Printf("Error retrieving all deck lists that feature cards w/ ID %v. Err: %v", cardIDs, err)
		return nil, &model.APIError{Message: "Could not get deck lists."}
	} else {
		dl := []model.DeckList{}
		if err := cursor.All(ctx, &dl); err != nil {
			log.Printf("Error retrieving all deck lists that feature cards w/ ID %v. Err: %v", cardIDs, err)
			return nil, &model.APIError{Message: "Could not get deck lists."}
		}

		return &dl, nil
	}
}

// Will update the database with a new traffic record.
func InsertTrafficData(ta model.TrafficAnalysis) *model.APIError {
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

func IsBlackListed(blackListType string, blackListPhrase string) (bool, *model.APIError) {
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

func GetCardOfTheDayForGivenDate(date string) (*string, *model.APIError) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	query := bson.M{"date": date, "version": 1}
	// select only these fields from collection
	opts := options.FindOne().SetProjection(
		bson.D{
			{Key: "cardID", Value: 1},
		},
	)

	var cotd model.CardOfTheDAy
	if err := cardOfTheDayCollection.FindOne(ctx, query, opts).Decode(&cotd); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil, nil
		}
		log.Printf("Error retrieving card of the day for given date: %s. Err: %s", date, err)
		return nil, &model.APIError{Message: "Could not get deck lists."}
	}

	return &cotd.CardID, nil
}
