package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetSkcSuggestionDBVersion() (string, error) {
	var commandResult bson.M
	var version string
	command := bson.D{{"serverStatus", 1}}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := skcSuggestionDB.RunCommand(ctx, command).Decode(&commandResult); err != nil {
		log.Println("Error getting SKC Suggestion DB version", err)
		return version, err
	} else {
		return version, nil
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
			log.Println(fmt.Sprintf("Error retrieving deck list w/ ID %s", deckID), err)
			return nil, &model.APIError{Message: "Requested deck list not found in DB."}
		} else {
			return &dl, nil
		}
	}
}

func InsertTrafficData(traffic model.TrafficAnalysis) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if res, err := skcSuggestionDB.Collection("trafficAnalysis").InsertOne(ctx, traffic); err != nil {
		log.Println("Error inserting traffic data into DB", err)
	} else {
		log.Println("Successfully inserted traffic data into DB, ID:", res.InsertedID)
	}
}
