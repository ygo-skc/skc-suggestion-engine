package skc_api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func ApiCall() {
	res, err := http.Get(SkcBaseUrl + CardInfoEndpoint)
	if err != nil {
		log.Fatalln("There was an error fetching info: ", err)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var cardInfoResponse CardInfoResponse
	json.Unmarshal(body, &cardInfoResponse)

	log.Println("Name of card: ", cardInfoResponse.CardName)
	log.Println("Name of card: ", cardInfoResponse.CardID)
	log.Println("Name of card: ", cardInfoResponse.CardAttribute)
	log.Println("Name of card: ", cardInfoResponse.CardEffect)
}
