package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/util"
)

func SubmitNewDeckList(res http.ResponseWriter, req *http.Request) {
	deckListName, encodedDeckList, tags := req.FormValue("name"), req.FormValue("list"), strings.Split(req.FormValue("tags"), ",")
	deckList := db.DeckList{Name: deckListName, ListContent: encodedDeckList, Tags: tags}
	log.Printf("Client submitting new deck with name {%s} and with list contents (in base64) {%s}", deckListName, encodedDeckList)

	res.Header().Add("Content-Type", "application/json") // prepping res headers

	// validate and handle validation error messages
	if err := util.V.Struct(deckList); err != nil {
		errMessages := []string{}
		for _, e := range err.(validator.ValidationErrors) {
			errMessages = append(errMessages, e.Translate(util.Translator))
		}

		message := strings.Join(errMessages, " ")
		log.Println("There were", len(errMessages), "errors while validating input. Errors:", message)

		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(util.APIError{Message: message})
		return
	}

	decodedListBytes, _ := base64.StdEncoding.DecodeString(encodedDeckList)
	decodedList := string(decodedListBytes) // decoded string of list contents

	var cardCopiesInDeck map[string]int
	var idsForCardsInDeckList []string
	var err util.APIError
	if cardCopiesInDeck, idsForCardsInDeckList, err = transformDeckListStringToMap(decodedList); err.Message != "" {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
	}

	var deckListDataFromDB map[string]db.Card
	if deckListDataFromDB, err = db.FindDesiredCardInDBUsingMultipleCardIDs(idsForCardsInDeckList); err.Message != "" {
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(err)
	}

	validateDeckList(cardCopiesInDeck, idsForCardsInDeckList, deckListDataFromDB)
	// Adding new deck list, fully validate before insertion
	db.InsertDeckList(deckList)
	json.NewEncoder(res).Encode(deckListDataFromDB)
}

// Transforms decoded deck list into a map that can be parsed easier.
// The map will use the cardID as key and number of copies in the deck as value.
func transformDeckListStringToMap(list string) (map[string]int, []string, util.APIError) {
	tokens := deckListCardAndQuantityRegex.FindAllString(list, -1)

	cardCopiesInDeck := map[string]int{}
	cards := []string{}
	for _, token := range tokens {
		splitToken := strings.Split(strings.ToLower(token), "x")
		quantity, _ := strconv.Atoi(splitToken[0])
		cardID := splitToken[1]

		if _, isPresent := cardCopiesInDeck[cardID]; isPresent {
			log.Printf("Deck list contains multiple instances of the same card {%s}.", cardID)
			return nil, nil, util.APIError{Message: "Deck list contains multiple instance of same card. Make sure a cardID appears only once in the deck list."}
		}
		cardCopiesInDeck[cardID] = quantity
		cards = append(cards, cardID)
	}

	log.Println("Decoded deck list, decoded contents:", cardCopiesInDeck)
	return cardCopiesInDeck, cards, util.APIError{}
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
