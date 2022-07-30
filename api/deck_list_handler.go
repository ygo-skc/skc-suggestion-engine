package api

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
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

	log.Printf("Client attempting to submit new deck with name {%s} and with list contents (in base64) {%s}", deckList.Name, deckList.ContentB64)

	res.Header().Add("Content-Type", "application/json") // prepping res headers

	// object validation
	if err := deckList.Validate(); err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	}

	decodedListBytes, _ := base64.StdEncoding.DecodeString(deckList.ContentB64)
	decodedList := string(decodedListBytes) // decoded string of list contents

	var deckListBreakdown model.DeckListBreakdown
	if dlb, err := getBreakdown(decodedList); err != nil {
		if err.Message == "Could not transform to map" {
			res.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(res).Encode(err)
		} else if err.Message == "Could not access DB" {
			res.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(res).Encode(err)
		}
	} else {
		deckListBreakdown = *dlb
	}

	deckListBreakdown.Sort()

	if err := deckListBreakdown.Validate(); err.Message != "" {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	}

	// Adding new deck list, fully validate before insertion
	deckList.ContentB64 = base64.StdEncoding.EncodeToString([]byte(deckListBreakdown.ListStringCleanup()))
	deckList.UniqueCards = deckListBreakdown.CardIDs
	deckList.NumMainDeckCards = deckListBreakdown.NumMainDeckCards
	deckList.NumExtraDeckCards = deckListBreakdown.NumExtraDeckCards
	db.InsertDeckList(deckList)
	json.NewEncoder(res).Encode(model.Success{Message: "Successfully inserted new deck list: " + deckList.Name})
}

func getBreakdown(dl string) (*model.DeckListBreakdown, *model.APIError) {
	var dlb model.DeckListBreakdown
	var err model.APIError

	if dlb, err = transformDeckListStringToMap(dl); err.Message != "" {
		return nil, &model.APIError{Message: "Could not transform to map"}
	}

	var allCards model.DeckListContents
	if allCards, err = db.FindDesiredCardInDBUsingMultipleCardIDs(dlb.CardIDs); err.Message != "" {
		return nil, &model.APIError{Message: "Could not access DB"}
	}

	dlb.AllCards = allCards
	return &dlb, nil
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

func getDeckList(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	deckID := pathVars["deckID"]
	log.Println("Getting content for deck w/ ID:", deckID)

	res.Header().Add("Content-Type", "application/json") // prepping res headers

	var deckList *model.DeckList
	var err *model.APIError
	if deckList, err = db.GetDeckList(deckID); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(err)
		return
	}

	decodedListBytes, _ := base64.StdEncoding.DecodeString(deckList.ContentB64)
	decodedList := string(decodedListBytes) // decoded string of list contents

	var deckListBreakdown model.DeckListBreakdown
	if dlb, err := getBreakdown(decodedList); err != nil {
		if err.Message == "Could not transform to map" {
			res.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(res).Encode(err)
		} else if err.Message == "Could not access DB" {
			res.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(res).Encode(err)
		}
	} else {
		deckListBreakdown = *dlb
	}

	deckListBreakdown.Sort()
	mainDeckContent := make([]model.Content, 0, len(deckListBreakdown.MainDeck))
	for _, card := range deckListBreakdown.MainDeck {
		mainDeckContent = append(mainDeckContent, model.Content{Card: card, Quantity: deckListBreakdown.CardQuantity[card.CardID]})
	}
	deckList.MainDeck = &mainDeckContent

	extraDeck := make([]model.Content, 0, len(deckListBreakdown.ExtraDeck))
	for _, card := range deckListBreakdown.ExtraDeck {
		extraDeck = append(extraDeck, model.Content{Card: card, Quantity: deckListBreakdown.CardQuantity[card.CardID]})
	}
	deckList.ExtraDeck = &extraDeck

	log.Printf("Successfully retrieved deck list. Name {%s} and encoded deck list content {%s}. This deck list has {%d} main deck cards and {%d} extra deck cards.", deckList.Name, deckList.ContentB64, deckList.NumMainDeckCards, deckList.NumExtraDeckCards)
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(deckList)
}
