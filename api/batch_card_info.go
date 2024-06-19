package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

func getBatchCardInfo(res http.ResponseWriter, req *http.Request) {
	logger, ctx := util.NewRequestSetup(context.Background(), "batch card info")
	logger.Info("Getting batch card info")

	// deserialize body
	var reqBody model.BatchCardIDs
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		logger.Error(fmt.Sprintf("Error occurred while reading the request body. Error %s", err))
		model.HandleServerResponse(model.APIError{Message: "Body could not be deserialized.", StatusCode: http.StatusBadRequest}, res)
		return
	}

	// validate body
	if err := validation.ValidateBatchCardIDs(reqBody); err != nil {
		err.HandleServerResponse(res)
		return
	}

	batchCardInfo := model.BatchCardData[model.CardIDs]{CardInfo: model.CardDataMap{}, UnknownResources: model.CardIDs{}}
	if len(reqBody.CardIDs) != 0 {
		// get card details
		var err *model.APIError
		if batchCardInfo, err = skcDBInterface.GetDesiredCardInDBUsingMultipleCardIDs(ctx, reqBody.CardIDs); err != nil {
			err.HandleServerResponse(res)
			return
		} else {
			if len(batchCardInfo.UnknownResources) > 0 {
				logger.Warn(fmt.Sprintf("Following card IDs are not valid (no card data found in DB). IDs: %v", batchCardInfo.UnknownResources))
			}
		}
	} else {
		logger.Info("Nothing to process - missing cardID data")
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(batchCardInfo)
}
