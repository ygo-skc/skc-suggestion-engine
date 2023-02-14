package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
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
	if s, err := skcDBInterface.FindOccurrenceOfCardNameInAllCardEffect(support.Card.CardName, cardID); err != nil {
		res.WriteHeader(err.StatusCode)
		json.NewEncoder(res).Encode(err)
	} else {
		support.ReferencedBy, support.MaterialFor = buildSupport(support.Card.CardName, s)

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(support)
	}
}

func buildSupport(cardName string, c *[]model.Card) (*[]model.Card, *[]model.Card) {
	support := make([]model.Card, 0)
	materialFor := make([]model.Card, 0)

	for _, card := range *c {
		tokens := quotedStringRegex.FindAllString(card.GetPotentialMaterialsAsString(), -1)
		if card.IsExtraDeckMonster() && isCardAMaterialForReference(tokens, cardName) {
			materialFor = append(materialFor, card)
		} else {
			support = append(support, card)
		}
	}

	return &support, &materialFor
}

func isCardAMaterialForReference(tokens []model.QuotedToken, cardName string) bool {
	isMaterialFor := false

	for _, token := range tokens {
		util.CleanupToken(&token)

		if cardName == token {
			isMaterialFor = true
			break
		}
	}

	return isMaterialFor
}
