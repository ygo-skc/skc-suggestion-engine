package db

import (
	"fmt"
	"log"
	"strings"

	"github.com/ygo-skc/skc-suggestion-engine/contracts"
)

func GetMaterialSuggestionsForCard(cardID string) error {
	desiredCard := FindDesiredCardInDB(cardID)

	materials := strings.SplitAfter(desiredCard.CardEffect, "\n")
	if len(materials) < 2 {
		return fmt.Errorf("could not determine materials")
	}
	log.Println(materials[0])

	return nil
}

func FindDesiredCardInDB(cardID string) contracts.Card {
	var card contracts.Card
	err := SKCDBConn.QueryRow("SELECT card_number, card_name, card_effect FROM cards WHERE card_number = ?", cardID).Scan(&card.CardID, &card.CardName, &card.CardEffect)

	if err != nil {
		log.Fatalln("Error occurred while fetching info for card w/ ID: ", cardID, err)
	}

	return card
}
