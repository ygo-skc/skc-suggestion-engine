package db

import "log"

func InsertDeckList(deckList DeckList) {
	if res, err := skcSuggestionEngineDeckListCollection.InsertOne(ctx, deckList); err != nil {
		log.Println("Error inserting new deck list to DB", err)
	} else {
		log.Println("Successfully inserted new deck list, ID:", res.InsertedID)
	}
}
