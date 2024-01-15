package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ygo-skc/skc-suggestion-engine/model"
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

	// TODO: validate body

	// get card details
	if cardData, err := skcDBInterface.FindDesiredCardInDBUsingMultipleCardIDs(reqBody.CardIDs); err != nil {
		err.HandleServerResponse(res)
	} else {
		batchCardInfo := model.BatchCardInfo{CardInfo: cardData, InvalidCardIDs: cardData.FindMissingIDs(reqBody.CardIDs)}

		if len(batchCardInfo.InvalidCardIDs) > 0 {
			log.Printf("Following card IDs are not valid (no card data found in DB). IDs: %v", batchCardInfo.InvalidCardIDs)
		}

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(batchCardInfo)
	}
}
