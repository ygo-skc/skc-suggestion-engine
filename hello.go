package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type CardInfoResponse struct {
	CardID        string `json:"cardID"`
	CardName      string `json:"cardName"`
	CardColor     string `json:"cardColor"`
	CardAttribute string `json:"cardAttribute"`
	CardEffect    string `json:"cardEffect"`
}

func main() {
	res, err := http.Get("https://skc-ygo-api.com/api/v1/card/67288539?allInfo=true")
	if err != nil {
		log.Fatalln("There was an error fetching info: ", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	var cardInfoResponse CardInfoResponse
	json.Unmarshal(body, &cardInfoResponse)

	fmt.Println("Name of card: ", cardInfoResponse.CardName)
	fmt.Println("Name of card: ", cardInfoResponse.CardID)
	fmt.Println("Name of card: ", cardInfoResponse.CardAttribute)
	fmt.Println("Name of card: ", cardInfoResponse.CardEffect)
}
