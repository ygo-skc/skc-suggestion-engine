package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

func getCardSupportHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]
	log.Printf("Getting cards that support card w/ ID: %s", cardID)

	support := model.CardSupport{ReferencedBy: []model.Card{}, MaterialFor: []model.Card{}}
	if cardToGetSupportFor, err := skcDBInterface.GetDesiredCardInDBUsingID(cardID); err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		support.Card = cardToGetSupportFor
	}

	// get support
	if s, err := skcDBInterface.GetOccurrenceOfCardNameInAllCardEffect(support.Card.CardName, cardID); err != nil {
		err.HandleServerResponse(res)
		return
	} else if len(s) == 0 {
		log.Println("No support found")
	} else {
		support.ReferencedBy, support.MaterialFor = determineSupportCards(support.Card, s)
		log.Printf("%s has %d cards that directly reference it (excluding cards referencing it as a material)", cardID, len(support.ReferencedBy))
		log.Printf("%s can be used as a material for %d cards", cardID, len(support.MaterialFor))
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(support)
}

// Iterates over a list of support cards and attempts to determine if subject is found in material clause or within the body of the reference.
// If the name is found in the material clause, we can assume the subject is a required or optional summoning material - otherwise its a support card.
func determineSupportCards(subject model.Card, references []model.Card) ([]model.Card, []model.Card) {
	referencedBy := []model.Card{}
	materialFor := []model.Card{}

	for _, reference := range references {
		materialString := reference.GetPotentialMaterialsAsString()
		materialStringTokens := quotedStringRegex.FindAllString(materialString, -1)

		remainingEffect := strings.Replace(reference.CardEffect, materialString, "", -1) // effect without materials
		remainingEffectTokens := quotedStringRegex.FindAllString(remainingEffect, -1)

		if reference.IsExtraDeckMonster() && subject.IsCardNameInTokens(materialStringTokens) {
			materialFor = append(materialFor, reference)
		}

		if subject.IsCardNameInTokens(remainingEffectTokens) {
			referencedBy = append(referencedBy, reference)
		}
	}

	return referencedBy, materialFor
}

func getBatchSupportHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Batch card suggestions requested")

	// TODO: below 3 conditions can be put in a method as they are shared in suggestions and support handler
	// deserialize body
	var reqBody model.BatchCardIDs
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		log.Printf("Error occurred while reading batch suggestions request body: Error %s", err)
		model.HandleServerResponse(model.APIError{Message: "Body could not be deserialized", StatusCode: http.StatusBadRequest}, res)
		return
	}

	// validate body
	if err := validation.ValidateBatchCardIDs(reqBody); err != nil {
		err.HandleServerResponse(res)
		return
	}

	if len(reqBody.CardIDs) == 0 {
		log.Println("Nothing to process - missing cardID data")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(noBatchSuggestions)
		return
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode("TBD")
}
