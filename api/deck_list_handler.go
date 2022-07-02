package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ygo-skc/skc-suggestion-engine/db"
)

func SubmitNewDeckList(res http.ResponseWriter, req *http.Request) {
	name, encodedList := req.FormValue("name"), req.FormValue("list")
	log.Println("Creating new deck list named", name, "and list contents (in base64)", encodedList)

	res.Header().Add("Content-Type", "application/json") // prepping res headers

	var decodedList string
	if decodedListBytes, err := base64.StdEncoding.DecodeString(encodedList); err != nil {
		log.Println("Could not decode card list input from user. Is it in base64? String causing issues:", encodedList, ". Error", err)

		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(APIError{Message: "Deck list not encoded correctly."})
		return
	} else {
		decodedList = string(decodedListBytes)
	}

	if cardCopiesInDeck, idsForCardsInDeckList, err := transformDeckListStringToMap(decodedList); err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(APIError{Message: "Deck list contains multiple instance of same card. Make sure each row contains a unique cardID."})
	} else {
		if deckListDataFromDB, err := db.FindDesiredCardInDBUsingMultipleCardIDs(idsForCardsInDeckList); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(res).Encode(APIError{Message: "Error occurred while validating deck list."})
		} else {
			validateDeckList(cardCopiesInDeck, idsForCardsInDeckList, deckListDataFromDB)
			// Adding new deck list, fully validate before insertion
			db.InsertDeckList(db.DeckList{Name: name, ListContent: encodedList})
			json.NewEncoder(res).Encode(deckListDataFromDB)
		}
	}
}

// Transforms decoded deck list into a map that can be parsed easier.
// The map will use the cardID as key and number of copies in the deck as value.
func transformDeckListStringToMap(list string) (map[string]int, []string, error) {
	tokens := deckListCardAndQuantityRegex.FindAllString(list, -1)

	cardCopiesInDeck := map[string]int{}
	cards := []string{}
	for _, token := range tokens {
		splitToken := strings.Split(strings.ToLower(token), "x")
		quantity, _ := strconv.Atoi(splitToken[0])
		cardID := splitToken[1]

		if _, isPresent := cardCopiesInDeck[cardID]; isPresent {
			log.Println("Deck list contains same cardID -", cardID, "- in multiple rows")
			return nil, nil, errors.New("422")
		}
		cardCopiesInDeck[cardID] = quantity
		cards = append(cards, cardID)
	}

	log.Println("Parsed decoded deck list", cardCopiesInDeck)
	return cardCopiesInDeck, cards, nil
}

func validateDeckList(cardCopiesInDeck map[string]int, idsForCardsInDeckList []string, deckListDataFromDB map[string]db.Card) error {
	invalidIDs := []string{}
	for _, cardID := range idsForCardsInDeckList {
		if _, isPresent := deckListDataFromDB[cardID]; !isPresent {
			invalidIDs = append(invalidIDs, cardID)
		}
	}

	if len(invalidIDs) > 0 {
		log.Println("Deck list contains card(s) that were not found in skc DB. All cards not found in DB:", invalidIDs)
		return errors.New("422")
	}

	return nil
}
