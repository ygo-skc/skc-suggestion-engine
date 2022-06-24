package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ygo-skc/skc-suggestion-engine/contracts"
)

var (
	SKCDBConn *sql.DB
)

func EstablishSKCDBConn() {
	var err error
	SKCDBConn, err = sql.Open("mysql", "root@/skc_api_db")

	if err != nil {
		log.Fatalln("Error occurred while trying to establish DB connection: ", err)
	}
}

func getAllCards() {
	rows, err := SKCDBConn.Query("select card_number, card_name from cards")

	if err != nil {
		log.Fatalln("Error occurred while trying to query all cards: ", err)
	}

	defer rows.Close()

	for rows.Next() {
		var card contracts.Card
		err := rows.Scan(&card.CardID, &card.CardName)

		if err != nil {
			log.Fatalln("Error occurred scanning row: ", err)
		}

		fmt.Println(card)
	}
}
