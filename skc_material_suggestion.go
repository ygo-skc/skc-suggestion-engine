package main

import (
	"fmt"
	"log"
)

func getMaterialSuggestions(cardID string) {
	rows, err := SKCDBConn.Query("SELECT card_number, card_name, card_effect FROM cards WHERE card_number = ?", cardID)

	if err != nil {
		log.Fatalln("Error occurred while fetching info for card w/ ID: ", cardID, err)
	}

	defer rows.Close()

	for rows.Next() {
		var card Card
		err := rows.Scan(&card.CardID, &card.CardName, &card.CardID)

		if err != nil {
			log.Fatalln("Error occurred scanning row: ", err)
		}

		fmt.Println(card)
	}
}
