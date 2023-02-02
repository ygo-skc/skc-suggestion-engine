package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func getCardSupportHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]
	log.Printf("Getting associated support cards for card w/ ID: %s", cardID)

	if cardToGetSupportFor, err := skcDBInterface.FindDesiredCardInDBUsingID(cardID); err != nil {
		res.WriteHeader(err.StatusCode)
		json.NewEncoder(res).Encode(err)
	} else {
		if err := getCardSupport(cardToGetSupportFor.CardName); err != nil {
			res.WriteHeader(err.StatusCode)
			json.NewEncoder(res).Encode(err)
		}
	}
}

func getCardSupport(cardName string) *model.APIError {
	log.Println(cardName)
	if results, err := skcDBInterface.FindOccurrenceOfCardNameInAllCardEffect(cardName); err != nil {
		return err
	} else {
		log.Printf("%v", results)
	}

	return nil
}
