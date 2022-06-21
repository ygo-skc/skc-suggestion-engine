package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// ApiCall()
	SkcDBConn()
}

func ApiCall() {
	res, err := http.Get(SkcBaseUrl + CardInfoEndpoint)
	if err != nil {
		log.Fatalln("There was an error fetching info: ", err)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var cardInfoResponse CardInfoResponse
	json.Unmarshal(body, &cardInfoResponse)

	fmt.Println("Name of card: ", cardInfoResponse.CardName)
	fmt.Println("Name of card: ", cardInfoResponse.CardID)
	fmt.Println("Name of card: ", cardInfoResponse.CardAttribute)
	fmt.Println("Name of card: ", cardInfoResponse.CardEffect)
}

type Card struct {
	CardName string `db:"card_name"`
}

func SkcDBConn() {
	db, err := sql.Open("mysql", "root@/skc_api_db")

	if err != nil {
		log.Fatalln("Error occurred while trying to establish DB connection", err)
	}

	rows, err := db.Query("select card_name from cards")

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var card Card
		err := rows.Scan(&card.CardName)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v\n", card)
	}

	defer db.Close()
}
