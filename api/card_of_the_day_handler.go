package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func getCardOfTheDay(res http.ResponseWriter, req *http.Request) {
	date := time.Now()
	log.Printf("Fetching card of the day - todays date %s", date.Format("2006-01-02"))

	if randomCardId, err := skcDBInterface.GetRandomCard(); err != nil {
		res.WriteHeader(err.StatusCode)
		json.NewEncoder(res).Encode(err)
		return
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(randomCardId)
	}
}
