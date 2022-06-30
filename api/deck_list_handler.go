package api

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ygo-skc/skc-suggestion-engine/db"
)

func SubmitNewDeckList(res http.ResponseWriter, req *http.Request) {
	name, list := req.FormValue("name"), req.FormValue("list")
	log.Println("Creating new deck list named", name, "and list contents (in base64)", list)

	res.Header().Add("Content-Type", "application/json") // prepping res headers

	if decodedList, err := base64.StdEncoding.DecodeString(list); err != nil {
		log.Println("Could not decode card list input from user. Is it in base64? String causing issues:", list, ". Error", err)

		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(APIError{Message: "Deck list not encoded correctly."})
		return
	} else {
		list = string(decodedList)
	}

	_, cardIDList := transformDeckListStringToMap(list)

	if cardData, err := db.FindDesiredCardInDBUsingMultipleCardIDs(cardIDList); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(APIError{Message: "Error occurred while validating deck list."})
	} else {
		json.NewEncoder(res).Encode(cardData)
	}
}

// Transforms decoded deck list into a map that can be parsed easier.
// The map will use the cardID as key and number of copies in the deck as value.
func transformDeckListStringToMap(list string) (map[string]int, []string) {
	tokens := deckListCardAndQuantityRegex.FindAllString(list, -1)

	deckList := map[string]int{}
	cards := []string{}
	for _, token := range tokens {
		splitToken := strings.Split(strings.ToLower(token), "x")
		quantity, _ := strconv.Atoi(splitToken[0])
		cardID := splitToken[1]

		deckList[cardID] = quantity
		cards = append(cards, cardID)
	}

	log.Println("Parsed decoded deck list", deckList)
	return deckList, cards
}
