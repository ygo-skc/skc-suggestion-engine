package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Retrieves the version number of the SKC Suggestion DB or throws an error if an exception occurs.
func GetSKCSuggestionDBVersion() (string, error) {
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

	if res, err := skcSuggestionDB.Collection("deckLists").InsertOne(ctx, deckList); err != nil {
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
		if err := skcSuggestionDB.Collection("deckLists").FindOne(ctx, bson.M{"_id": objectId}).Decode(&dl); err != nil {
			log.Printf("Error retrieving deck list w/ ID %s. Err: %v", deckID, err)
			return nil, &model.APIError{Message: "Requested deck list not found in DB."}
		} else {
			return &dl, nil
		}
	}
}

func GetDecksThatFeatureCards(cardIDs []string) (*[]model.DeckList, *model.APIError) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	opts := options.Find().SetProjection(bson.D{{Key: "name", Value: 1}, {Key: "videoUrl", Value: 1}, {Key: "createdAt", Value: 1}, {Key: "updatedAt", Value: 1}})

	if cursor, err := skcSuggestionDB.Collection("deckLists").Find(ctx, bson.M{"uniqueCards": bson.M{"$in": cardIDs}}, opts); err != nil {
		log.Printf("Error retrieving all deck lists that feature cards w/ ID %v. Err: %v", cardIDs, err)
		return nil, &model.APIError{Message: "Could not get deck lists."}
	} else {
		var dl []model.DeckList
		if err := cursor.All(ctx, &dl); err != nil {
			log.Printf("Error retrieving all deck lists that feature cards w/ ID %v. Err: %v", cardIDs, err)
			return nil, &model.APIError{Message: "Could not get deck lists."}
		}

		return &dl, nil
	}
}

func InsertTrafficData(ta model.TrafficAnalysis) *model.APIError {
	log.Printf("Inserting traffic data for resource %v and system %v.", ta.ResourceUtilized, ta.Source)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if res, err := skcSuggestionDB.Collection("trafficAnalysis").InsertOne(ctx, ta); err != nil {
		log.Println("Error inserting traffic data into DB", err)

		return &model.APIError{Message: "Error occurred while attempting to insert new traffic data."}
	} else {
		log.Println("Successfully inserted traffic data into DB, ID:", res.InsertedID)
	}

	return nil
}
