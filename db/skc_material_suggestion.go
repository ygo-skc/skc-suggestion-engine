package db

import (
	"log"

	"github.com/ygo-skc/skc-suggestion-engine/contracts"
)

const (
	queryCardUsingCardID   string = "SELECT card_number, card_name, card_effect FROM cards WHERE card_number = ?"
	queryCardUsingCardName string = "SELECT card_number, card_name, card_effect FROM cards WHERE card_name = ?"
)

func FindDesiredCardInDBUsingID(cardID string) (contracts.Card, error) {
	var card contracts.Card

	if err := skcDBConn.QueryRow(queryCardUsingCardID, cardID).Scan(&card.CardID, &card.CardName, &card.CardEffect); err != nil {
		return card, err
	}

	return card, nil
}

func FindDesiredCardInDBUsingName(cardName string) contracts.Card {
	var card contracts.Card
	err := skcDBConn.QueryRow(queryCardUsingCardName, cardName).Scan(&card.CardID, &card.CardName, &card.CardEffect)

	if err != nil { // TODO: This should be updated to bubble up the err
		log.Fatalln("Error occurred while fetching info for card w/ name: ", cardName, err)
	}

	return card
}
