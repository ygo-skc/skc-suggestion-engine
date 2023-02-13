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

	var support model.CardSupport
	if cardToGetSupportFor, err := skcDBInterface.FindDesiredCardInDBUsingID(cardID); err != nil {
		res.WriteHeader(err.StatusCode)
		json.NewEncoder(res).Encode(err)
		return
	} else {
		support.Card = cardToGetSupportFor
	}

	// get support
	if s, err := skcDBInterface.FindOccurrenceOfCardNameInAllCardEffect(support.Card.CardName); err != nil {
		res.WriteHeader(err.StatusCode)
		json.NewEncoder(res).Encode(err)
	} else {
		references := make([]model.CardReference, len(*s))

		for ind, card := range *s {
			references[ind] = model.CardReference{Card: card}
		}

		support.Support = &references

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(support)
	}
}
