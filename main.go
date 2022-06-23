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
	GetMaterialSuggestions("35809262")
	SetupMultiplexer()
}

func SetupMultiplexer() {
	http.HandleFunc("/api/v1/materials/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "Supp")
	})

	if err := http.ListenAndServe("localhost:8081", nil); err != nil {
		log.Fatalln("There was an error starting server: ", err)
	}
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
