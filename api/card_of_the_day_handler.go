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

	if cardID, err := skcSuggestionEngineDBInterface.GetCardOfTheDay(date); cardID == nil {
		if err := fetchNewCardOfTheDayAndPersist(&cardOfTheDay); err != nil {
			err.HandleServerResponse(res)
			return
		}
	} else if err != nil {
		err.HandleServerResponse(res)
	} else {
		log.Printf("Existing card of the day for %s found! Card of the day: %s", date, *cardID)
		cardOfTheDay.CardID = *cardID
	}

	log.Println("Fetching card of the day information.")
	if card, err := skcDBInterface.FindDesiredCardInDBUsingID(cardOfTheDay.CardID); err != nil {
		e := &model.APIError{StatusCode: http.StatusInternalServerError, Message: "An error occurred fetching card of the day details."}
		e.HandleServerResponse(res)
		return
	} else {
		cardOfTheDay.Card = card
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(cardOfTheDay)
}

func fetchNewCardOfTheDayAndPersist(cotd *model.CardOfTheDay) *model.APIError {
	log.Printf("There was no card of the day found for %s, fetching random card from DB.", cotd.Date)
	e := &model.APIError{StatusCode: http.StatusInternalServerError, Message: "An error occurred fetching new card of the day."}

	if randomCardId, err := skcDBInterface.GetRandomCard(); err != nil {
		return e
	} else {
		cotd.CardID = randomCardId
	}

	if err := skcSuggestionEngineDBInterface.InsertCardOfTheDay(*cotd); err != nil {
		return e
	}

	return nil
}
