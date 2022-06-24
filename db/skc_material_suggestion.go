package db

import (
	"log"

	"github.com/ygo-skc/skc-suggestion-engine/contracts"
)

func FindDesiredCardInDB(cardID string) contracts.Card {
	var card contracts.Card
	err := SKCDBConn.QueryRow("SELECT card_number, card_name, card_effect FROM cards WHERE card_number = ?", cardID).Scan(&card.CardID, &card.CardName, &card.CardEffect)

	if err != nil {
		log.Fatalln("Error occurred while fetching info for card w/ ID: ", cardID, err)
	}

	return card
}

func FindDesiredCardInDBUsingName(cardName string) contracts.Card {
	var card contracts.Card
	err := SKCDBConn.QueryRow("SELECT card_number, card_name, card_effect FROM cards WHERE card_name = ?", cardName).Scan(&card.CardID, &card.CardName, &card.CardEffect)

	if err != nil { // TODO: This should be updated to bubble up the err
		log.Fatalln("Error occurred while fetching info for card w/ name: ", cardName, err)
	}

	return card
}
