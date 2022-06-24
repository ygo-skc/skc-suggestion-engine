package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// ApiCall()
	EstablishSKCDBConn()
	// getAllCards()
	GetMaterialSuggestionsForCard("35809262")
	SetupMultiplexer()
}

func SetupMultiplexer() {
	http.HandleFunc("/api/v1/suggestions/materials", GetMaterialSuggestions)

	if err := http.ListenAndServe("localhost:9000", nil); err != nil {
		log.Fatalln("There was an error starting server: ", err)
	}
}

func GetMaterialSuggestions(res http.ResponseWriter, req *http.Request) {
	cards := []Card{
		{CardName: "Elemental HERO Avian"},
		{CardName: "Elemental HERO Burstinatrix"},
	}

	res.Header().Add("Content-Type", "application/json")
	json.NewEncoder(res).Encode(cards)
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
