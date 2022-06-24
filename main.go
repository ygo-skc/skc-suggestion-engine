package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ygo-skc/skc-suggestion-engine/api"
	"github.com/ygo-skc/skc-suggestion-engine/contracts"
	"github.com/ygo-skc/skc-suggestion-engine/db"
)

func main() {
	// ApiCall()
	db.EstablishSKCDBConn()
	api.SetupMultiplexer()
	// getAllCards()
}

func ApiCall() {
	res, err := http.Get(contracts.SkcBaseUrl + contracts.CardInfoEndpoint)
	if err != nil {
		log.Fatalln("There was an error fetching info: ", err)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var cardInfoResponse contracts.CardInfoResponse
	json.Unmarshal(body, &cardInfoResponse)

	fmt.Println("Name of card: ", cardInfoResponse.CardName)
	fmt.Println("Name of card: ", cardInfoResponse.CardID)
	fmt.Println("Name of card: ", cardInfoResponse.CardAttribute)
	fmt.Println("Name of card: ", cardInfoResponse.CardEffect)
}
