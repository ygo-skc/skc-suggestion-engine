package api

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
)

func submitNewDeckList(res http.ResponseWriter, req *http.Request) {
	deckListName, encodedDeckList, tags := req.FormValue("name"), req.FormValue("list"), strings.Split(req.FormValue("tags"), ",")
	deckList := model.DeckList{Name: deckListName, ListContent: encodedDeckList, Tags: tags}
	log.Printf("Client submitting new deck with name {%s} and with list contents (in base64) {%s}", deckListName, encodedDeckList)

	res.Header().Add("Content-Type", "application/json") // prepping res headers

	if err := deckList.Validate(); err.Message != "" {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	}

	decodedListBytes, _ := base64.StdEncoding.DecodeString(encodedDeckList)
	decodedList := string(decodedListBytes) // decoded string of list contents

	var deckListBreakdown model.DeckListBreakdown
	var err util.APIError
	if deckListBreakdown, err = transformDeckListStringToMap(decodedList); err.Message != "" {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
	}

	var deckListContents model.DeckListContents
	if deckListContents, err = db.FindDesiredCardInDBUsingMultipleCardIDs(deckListBreakdown.CardIDs); err.Message != "" {
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(err)
	}

	if err := deckListContents.Validate(deckListBreakdown.CardQuantity, deckListBreakdown.CardIDs); err.Message != "" {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	}

	// Adding new deck list, fully validate before insertion
	db.InsertDeckList(deckList)
	json.NewEncoder(res).Encode(deckListContents)
}

// Transforms decoded deck list into a map that can be parsed easier.
// The map will use the cardID as key and number of copies in the deck as value.
func transformDeckListStringToMap(list string) (model.DeckListBreakdown, util.APIError) {
	tokens := deckListCardAndQuantityRegex.FindAllString(list, -1)

	cardCopiesInDeck := map[string]int{}
	cards := []string{}
	for _, token := range tokens {
		splitToken := strings.Split(strings.ToLower(token), "x")
		quantity, _ := strconv.Atoi(splitToken[0])
		cardID := splitToken[1]

		if _, isPresent := cardCopiesInDeck[cardID]; isPresent {
			log.Printf("Deck list contains multiple instances of the same card {%s}.", cardID)
			return model.DeckListBreakdown{}, util.APIError{Message: "Deck list contains multiple instance of same card. Make sure a cardID appears only once in the deck list."}
		}
		cardCopiesInDeck[cardID] = quantity
		cards = append(cards, cardID)
	}

	log.Println("Decoded deck list contents:", cardCopiesInDeck)
	return model.DeckListBreakdown{CardQuantity: cardCopiesInDeck, CardIDs: cards}, util.APIError{}
}
