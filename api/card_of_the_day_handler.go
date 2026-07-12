package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	cModel "github.com/ygo-skc/skc-go/common/v2/model"
	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	cardOfTheDayOp = "Card of The Day"
)

func getCardOfTheDay(res http.ResponseWriter, req *http.Request) {
	logger, ctx := cUtil.InitRequest(context.Background(), apiName, cardOfTheDayOp)

	cardOfTheDay := model.CardOfTheDay{Date: time.Now().In(chicagoLocation).Format("2006-01-02"), Version: 1}
	logger.Info("Fetching card of the day", "date", cardOfTheDay.Date)

	if cardID, err := skcSuggestionEngineDBInterface.GetCardOfTheDay(ctx, cardOfTheDay.Date, cardOfTheDay.Version); cardID == nil {
		if err := fetchNewCardOfTheDayAndPersist(ctx, &cardOfTheDay); err != nil {
			err.HandleServerResponse(res)
			return
		}
	} else if err != nil {
		err.HandleServerResponse(res)
	} else {
		logger.Warn("Existing card of the day exists", "cotd", *cardID)
		cardOfTheDay.CardID = *cardID
	}

	if card, err := downstream.YGO.CardService.GetCardByID(ctx, cardOfTheDay.CardID); err != nil {
		e := &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "An error occurred fetching card of the day details."}
		e.HandleServerResponse(res)
		return
	} else {
		cardOfTheDay.Card = *card
	}

	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(cardOfTheDay); err != nil {
		logger.Error("Could not encode card of the day response", "err", err, "card_of_the_day_id", cardOfTheDay.CardID, "date", cardOfTheDay.Date)
	}
}

func fetchNewCardOfTheDayAndPersist(ctx context.Context, cotd *model.CardOfTheDay) *cModel.APIError {
	logger := cUtil.RetrieveLogger(ctx)
	logger.Info("There was no COTD picked for today - getting random card")
	e := &cModel.APIError{StatusCode: http.StatusInternalServerError, Message: "An error occurred fetching new card of the day."}

	var err *cModel.APIError
	var previousCOTDData []string
	if previousCOTDData, err = skcSuggestionEngineDBInterface.GetHistoricalCardOfTheDayData(ctx, cotd.Version); err != nil {
		return e
	}

	logger.Warn("Ignoring cards that were previously COTD", "total_ignored", len(previousCOTDData))

	if randomCard, err := downstream.YGO.CardService.GetRandomCardProto(ctx, previousCOTDData); err != nil {
		return e
	} else {
		cotd.CardID = randomCard.ID
	}

	if err := skcSuggestionEngineDBInterface.InsertCardOfTheDay(ctx, *cotd); err != nil {
		return e
	}

	return nil
}
