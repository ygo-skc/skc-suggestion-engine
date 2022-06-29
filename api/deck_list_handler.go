package api

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	if _, err := transformDeckListStringToMap(list); err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(APIError{Message: "Decoded card list not formatted correctly."})
	}

	json.NewEncoder(res).Encode("good")
}

// Transforms decoded deck list into a map that can be parsed easier.
// The map will use the cardID as key and number of copies in the deck as value.
func transformDeckListStringToMap(list string) (map[string]int, error) {
	tokens := deckListCardAndQuantityRegex.FindAllString(list, -1)

	var deckList = map[string]int{}
	for _, token := range tokens {
		splitToken := strings.Split(strings.ToLower(token), "x")
		quantity, _ := strconv.Atoi(splitToken[0])
		cardID := splitToken[1]
		deckList[cardID] = quantity
	}

	log.Println("Parsed decoded deck list", deckList)
	return deckList, nil
}
