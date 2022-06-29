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
		json.NewEncoder(res).Encode(APIError{Message: "Card list input in not formatted correctly."})
		return
	} else {
		list = string(decodedList)
	}

	tokens := deckListCardAndQuantityRegex.FindAllString(list, -1)
	var deckList = map[string]int{}
	for _, token := range tokens {
		t := strings.Split(strings.ToLower(token), "x")
		if quantity, err := strconv.Atoi(t[0]); err != nil { // quantity string was not an int - this shouldn't happen as regex expects a digit
			log.Println("Could not convert string to int for quantity field. Err:", err)

			res.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(res).Encode(APIError{Message: "Decoded card list data not formatted correctly."})
		} else {
			cardID := t[1]
			deckList[cardID] = quantity
		}
	}
	log.Println(deckList)

	json.NewEncoder(res).Encode("good")
}
