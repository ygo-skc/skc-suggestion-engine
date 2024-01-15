package api

import (
	"encoding/json"
	"fmt"
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
		missingIDs := cardData.FindMissingIDs(reqBody.CardIDs)

		if len(missingIDs) > 0 {
			msg := fmt.Sprintf("Following card IDs are not valid (no card data found in DB). IDs: %v", missingIDs)
			log.Println(msg)

			model.HandleServerResponse(model.APIError{Message: msg, StatusCode: http.StatusNotFound}, res)
			return
		}

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(cardData)
	}
}
