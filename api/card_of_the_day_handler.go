package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
)

func getCardOfTheDay(res http.ResponseWriter, req *http.Request) {
	logger, ctx := util.NewRequestSetup(context.Background(), "card of the day")

	date := time.Now().In(chicagoLocation).Format("2006-01-02")
	cardOfTheDay := model.CardOfTheDay{Date: date, Version: 1}
	logger.Info(fmt.Sprintf("Fetching card of the day - todays date %s", date))

	if cardID, err := skcSuggestionEngineDBInterface.GetCardOfTheDay(date); cardID == nil {
		if err := fetchNewCardOfTheDayAndPersist(ctx, &cardOfTheDay); err != nil {
			err.HandleServerResponse(res)
			return
		}
	} else if err != nil {
		err.HandleServerResponse(res)
	} else {
		logger.Warn(fmt.Sprintf("Existing card of the day for %s found! Card of the day: %s", date, *cardID))
		cardOfTheDay.CardID = *cardID
	}

	if card, err := skcDBInterface.GetDesiredCardInDBUsingID(ctx, cardOfTheDay.CardID); err != nil {
		e := &model.APIError{StatusCode: http.StatusInternalServerError, Message: "An error occurred fetching card of the day details."}
		e.HandleServerResponse(res)
		return
	} else {
		cardOfTheDay.Card = card
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(cardOfTheDay)
}

func fetchNewCardOfTheDayAndPersist(ctx context.Context, cotd *model.CardOfTheDay) *model.APIError {
	util.Logger(ctx).Info(fmt.Sprintf("There was no card of the day found for %s, fetching random card from DB.", cotd.Date))
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
