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
	log.Printf("Getting cards that support card w/ ID: %s", cardID)

	support := model.CardSupport{ReferencedBy: []model.Card{}, MaterialFor: []model.Card{}}
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
		return
	} else if len(*s) == 0 {
		log.Println("No support found")
	} else {
		support.ReferencedBy, support.MaterialFor = determineSupportCards(support.Card.CardName, s)
		log.Printf("%s has %d cards that directly reference it (excluding cards referencing it as a material)", cardID, len(support.ReferencedBy))
		log.Printf("%s can be used as a material for %d cards", cardID, len(support.MaterialFor))
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(support)
}

// iterates over a list of cards and attempts to determine if given cardName is found in material clause or within the body of the card
// if the name is found in the material clause, we can assume the cardName is a required or optional summoning material - otherwise its a support card
func determineSupportCards(cardName string, c *[]model.Card) ([]model.Card, []model.Card) {
	referencedBy := []model.Card{}
	materialFor := []model.Card{}

	for _, card := range *c {
		tokens := quotedStringRegex.FindAllString(card.GetPotentialMaterialsAsString(), -1)
		if card.IsExtraDeckMonster() && isCardAMaterialForReference(tokens, cardName) {
			materialFor = append(materialFor, card)
		} else {
			referencedBy = append(referencedBy, card)
		}
	}

	return referencedBy, materialFor
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
