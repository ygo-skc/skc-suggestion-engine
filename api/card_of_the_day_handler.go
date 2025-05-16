package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cModel "github.com/ygo-skc/skc-go/common/model"
	"github.com/ygo-skc/skc-go/common/service"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func getCardOfTheDay(res http.ResponseWriter, req *http.Request) {
	logger, ctx := cUtil.NewRequestSetup(context.Background(), "card of the day")

	cardOfTheDay := model.CardOfTheDay{Date: time.Now().In(chicagoLocation).Format("2006-01-02"), Version: 1}
	logger.Info(fmt.Sprintf("Fetching card of the day - todays date %s", cardOfTheDay.Date))

	if cardID, err := skcSuggestionEngineDBInterface.GetCardOfTheDay(ctx, cardOfTheDay.Date, cardOfTheDay.Version); cardID == nil {
		if err := fetchNewCardOfTheDayAndPersist(ctx, &cardOfTheDay); err != nil {
			err.HandleServerResponse(res)
			return
		}
	} else if err != nil {
		err.HandleServerResponse(res)
	} else {
		logger.Warn(fmt.Sprintf("Existing card of the day found! COTD: %s", *cardID))
		cardOfTheDay.CardID = *cardID
	}

	if card, err := service.QueryCard(ctx, downstream.CardServiceClient, cardOfTheDay.CardID, cModel.YGOCardRESTFromPB); err != nil {
		e := &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "An error occurred fetching card of the day details."}
		e.HandleServerResponse(res)
		return
	} else {
		cardOfTheDay.Card = *card
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(cardOfTheDay)
}

func fetchNewCardOfTheDayAndPersist(ctx context.Context, cotd *model.CardOfTheDay) *cModel.APIError {
	logger := cUtil.LoggerFromContext(ctx)
	logger.Info("There was no COTD picked for today - getting random card")
	e := &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "An error occurred fetching new card of the day."}

	var err *cModel.APIError
	var previousCOTDData []string
	if previousCOTDData, err = skcSuggestionEngineDBInterface.GetHistoricalCardOfTheDayData(ctx, cotd.Version); err != nil {
		return e
	}

	logger.Warn(fmt.Sprintf("Ignoring cards that were previously COTD, total ignored: %d", len(previousCOTDData)))

	if randomCardId, err := skcDBInterface.GetRandomCard(ctx, previousCOTDData); err != nil {
		return e
	} else {
		cotd.CardID = randomCardId
	}

	if err := skcSuggestionEngineDBInterface.InsertCardOfTheDay(ctx, *cotd); err != nil {
		return e
	}

	return nil
}
