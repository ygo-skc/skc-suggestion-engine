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
	cardOfTheDay := model.CardOfTheDAy{Date: date, Version: 1}
	log.Printf("Fetching card of the day - todays date %s", date)

	if cardID, _ := skcSuggestionEngineDBInterface.GetCardOfTheDayForGivenDate(date); cardID == nil {
		log.Printf("There was no card of the day found for %s, fetching random card from DB.", date)
		if err := fetchNewCardOfTheDayAndPersist(&cardOfTheDay); err != nil {
			res.WriteHeader(err.StatusCode)
			json.NewEncoder(res).Encode(err)
			return
		}
	} else {
		log.Printf("Existing card of the day for %s found!", date)
		cardOfTheDay.CardID = *cardID
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(cardOfTheDay)
}

func fetchNewCardOfTheDayAndPersist(cotd *model.CardOfTheDAy) *model.APIError {
	if randomCardId, err := skcDBInterface.GetRandomCard(); err != nil {
		return err
	} else {
		cotd.CardID = randomCardId
	}

	return nil
}
