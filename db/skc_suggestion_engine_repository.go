package db

import (
	"context"
	"log"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/contract"
)

func InsertDeckList(deckList contract.DeckList) {
	deckList.CreatedAt = time.Now()
	deckList.UpdatedAt = deckList.CreatedAt

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if res, err := skcSuggestionEngineDeckListCollection.InsertOne(ctx, deckList); err != nil {
		log.Println("Error inserting new deck list to DB", err)
	} else {
		log.Println("Successfully inserted new deck list, ID:", res.InsertedID)
	}
}
