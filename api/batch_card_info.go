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
	var cardIDs []string
	if err := json.NewDecoder(req.Body).Decode(&cardIDs); err != nil {
		log.Printf("Error occurred while reading the request body. Error %s", err)

		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(model.APIError{Message: "Body could not be deserialized."})
		return
	}

	// TODO: validate body

	// get card details
	if cardData, err := skcDBInterface.FindDesiredCardInDBUsingMultipleCardIDs(cardIDs); err != nil {
		// return &model.APIError{Message: "Could not access DB"}
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(cardData)
	}
}
