package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func getCardOfTheDay(res http.ResponseWriter, req *http.Request) {
	date := time.Now().Format("2006-01-02")
	cardOfTheDay := model.CardOfTheDAy{Date: date, Version: 1}
	log.Printf("Fetching card of the day - todays date %s", date)

	if cardID, _ := db.GetCardOfTheDayForGivenDate(date); cardID == nil {
		log.Println("There was no card of the day selected for todays date, fetching random card.")
		if randomCardId, err := skcDBInterface.GetRandomCard(); err != nil {
			res.WriteHeader(err.StatusCode)
			json.NewEncoder(res).Encode(err)
			return
		} else {
			cardOfTheDay.CardID = randomCardId
		}
	} else {
		log.Println("Existing card of the day for todays date found!")
		cardOfTheDay.CardID = *cardID
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(cardOfTheDay)
}
