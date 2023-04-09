package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func getCardOfTheDay(res http.ResponseWriter, req *http.Request) {
	date := time.Now().Format("2006-01-02")
	cardOfTheDay := model.CardOfTheDay{Date: date, Version: 1}
	log.Printf("Fetching card of the day - todays date %s", date)

	if cardID, _ := skcSuggestionEngineDBInterface.GetCardOfTheDay(date); cardID == nil {
		if err := fetchNewCardOfTheDayAndPersist(&cardOfTheDay); err != nil {
			res.WriteHeader(err.StatusCode)
			json.NewEncoder(res).Encode(err)
			return
		}
	} else {
		log.Printf("Existing card of the day for %s found! Card of the day: %s", date, *cardID)
		cardOfTheDay.CardID = *cardID
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(cardOfTheDay)
}

func fetchNewCardOfTheDayAndPersist(cotd *model.CardOfTheDay) *model.APIError {
	log.Printf("There was no card of the day found for %s, fetching random card from DB.", cotd.Date)
	if randomCardId, err := skcDBInterface.GetRandomCard(); err != nil {
		return err
	} else {
		cotd.CardID = randomCardId
	}

	if err := skcSuggestionEngineDBInterface.InsertCardOfTheDay(*cotd); err != nil {
		return err
	}

	return nil
}
