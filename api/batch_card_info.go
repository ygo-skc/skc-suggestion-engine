package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

func getBatchCardInfo(res http.ResponseWriter, req *http.Request) {
	log.Println("Getting batch card info...")

	// deserialize body
	var reqBody model.BatchCardIDs
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		log.Printf("Error occurred while reading the request body. Error %s", err)
		model.HandleServerResponse(model.APIError{Message: "Body could not be deserialized.", StatusCode: http.StatusBadRequest}, res)
		return
	}

	// validate body
	if err := validation.ValidateBatchCardIDs(reqBody); err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	}

	var batchCardInfo model.BatchCardInfo
	if len(reqBody.CardIDs) == 0 {
		batchCardInfo = model.BatchCardInfo{CardInfo: model.CardDataMap{}, UnknownCardIDs: model.CardIDs{}}
	} else {
		// get card details
		if cardData, err := skcDBInterface.FindDesiredCardInDBUsingMultipleCardIDs(reqBody.CardIDs); err != nil {
			err.HandleServerResponse(res)
		} else {
			batchCardInfo = model.BatchCardInfo{CardInfo: cardData, UnknownCardIDs: cardData.FindMissingIDs(reqBody.CardIDs)}

			if len(batchCardInfo.UnknownCardIDs) > 0 {
				log.Printf("Following card IDs are not valid (no card data found in DB). IDs: %v", batchCardInfo.UnknownCardIDs)
			}
		}
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(batchCardInfo)
}
