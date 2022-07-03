package db

import (
	"log"
	"time"
)

func InsertDeckList(deckList DeckList) {
	deckList.CreatedAt = time.Now()
	deckList.UpdatedAt = deckList.CreatedAt

	if res, err := skcSuggestionEngineDeckListCollection.InsertOne(ctx, deckList); err != nil {
		log.Println("Error inserting new deck list to DB", err)
	} else {
		log.Println("Successfully inserted new deck list, ID:", res.InsertedID)
	}
}
