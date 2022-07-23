package api

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

func submitNewDeckList(res http.ResponseWriter, req *http.Request) {
	var deckList model.DeckList

	if b, err := ioutil.ReadAll(req.Body); err != nil {
		log.Println("Error occurred while reading the request body.")
	} else {
		json.Unmarshal(b, &deckList)
	}

	log.Printf("Client submitting new deck with name {%s} and with list contents (in base64) {%s}", deckList.Name, deckList.ListContent)

	res.Header().Add("Content-Type", "application/json") // prepping res headers

	if err := deckList.Validate(); err.Message != "" {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	}

	decodedListBytes, _ := base64.StdEncoding.DecodeString(deckList.ListContent)
	decodedList := string(decodedListBytes) // decoded string of list contents

	var deckListBreakdown model.DeckListBreakdown
	var err model.APIError
	if deckListBreakdown, err = transformDeckListStringToMap(decodedList); err.Message != "" {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	}

	var deckListContents model.DeckListContents
	if deckListContents, err = db.FindDesiredCardInDBUsingMultipleCardIDs(deckListBreakdown.CardIDs); err.Message != "" {
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(err)
		return
	}

	if err := deckListContents.Validate(deckListBreakdown.CardQuantity, deckListBreakdown.CardIDs); err.Message != "" {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	}

	// Adding new deck list, fully validate before insertion
	db.InsertDeckList(deckList)
	json.NewEncoder(res).Encode(model.Success{Message: "Successfully inserted new deck list: " + deckList.Name})
}

// Transforms decoded deck list into a map that can be parsed easier.
// The map will use the cardID as key and number of copies in the deck as value.
func transformDeckListStringToMap(list string) (model.DeckListBreakdown, model.APIError) {
	tokens := deckListCardAndQuantityRegex.FindAllString(list, -1)

	cardCopiesInDeck := map[string]int{}
	cards := []string{}
	for _, token := range tokens {
		splitToken := strings.Split(strings.ToLower(token), "x")
		quantity, _ := strconv.Atoi(splitToken[0])
		cardID := splitToken[1]

		if _, isPresent := cardCopiesInDeck[cardID]; isPresent {
			log.Printf("Deck list contains multiple instances of the same card {%s}.", cardID)
			return model.DeckListBreakdown{}, model.APIError{Message: "Deck list contains multiple instance of same card. Make sure a cardID appears only once in the deck list."}
		}
		cardCopiesInDeck[cardID] = quantity
		cards = append(cards, cardID)
	}

	return model.DeckListBreakdown{CardQuantity: cardCopiesInDeck, CardIDs: cards}, model.APIError{}
}
