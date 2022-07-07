package db

import (
	"context"
	"log"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func InsertDeckList(deckList model.DeckList) {
	deckList.CreatedAt = time.Now()
	deckList.UpdatedAt = deckList.CreatedAt

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if res, err := skcSuggestionEngineDeckListCollection.InsertOne(ctx, deckList); err != nil {
		log.Println("Error inserting new deck list into DB", err)
	} else {
		log.Println("Successfully inserted new deck list into DB, ID:", res.InsertedID)
	}
}
