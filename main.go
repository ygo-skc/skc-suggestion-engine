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
	getAllCards()
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
