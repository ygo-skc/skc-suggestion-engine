package main

import (
	"database/sql"
	"fmt"
	"log"
)

type Card struct {
	CardID     string `db:"card_number" json:"cardID"`
	CardName   string `db:"card_name" json:"cardName"`
	CardEffect string `db:"card_effect" json:"cardEffect"`
}

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
		var card Card
		err := rows.Scan(&card.CardID, &card.CardName)

		if err != nil {
			log.Fatalln("Error occurred scanning row: ", err)
		}

		fmt.Println(card)
	}
}
