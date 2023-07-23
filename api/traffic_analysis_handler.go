package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

// Endpoint will allow clients to submit traffic data to be saved in a MongoDB instance.
func submitNewTrafficDataHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Adding new traffic record...")

	// deserialize body
	var trafficData model.TrafficAnalysisInput
	if err := json.NewDecoder(req.Body).Decode(&trafficData); err != nil {
		log.Println("Error occurred while reading the request body.")

		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(model.APIError{Message: "Body could not be deserialized."})
		return
	}

	// validate body
	if err := trafficData.Validate(); err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	}

	// get IP number info
	var location model.Location
	if ipData, err := ipDB.Get_all(trafficData.IP); err != nil {
		log.Printf("Error getting info for IP address %s. Error %v", trafficData.IP, err)

		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(model.APIError{Message: "The IP provided was not found in the IP Database. Therefor, not storing traffic pattern."})
		return
	} else {
		location = model.Location{Zip: ipData.Zipcode, City: ipData.City, Country: ipData.Country_short}
	}

	// create traffic analysis object that will be inserted to DB
	userData := model.UserData{Location: location, IP: trafficData.IP}
	source := model.TrafficSource{SystemName: trafficData.Source.SystemName, Version: trafficData.Source.Version}
	trafficAnalysis := model.TrafficAnalysis{Timestamp: time.Now(), UserData: userData, ResourceUtilized: *trafficData.ResourceUtilized, Source: source}

	if err := skcSuggestionEngineDBInterface.InsertTrafficData(trafficAnalysis); err != nil {
		res.WriteHeader(err.StatusCode)
		json.NewEncoder(res).Encode(err)
		return
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(model.Success{Message: "Successfully inserted new traffic data."})
}

func trending(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	resource := strings.ToUpper(pathVars["resource"])
	log.Printf("Getting trending data for resource: %s", resource)

	c1, c2 := make(chan *model.APIError), make(chan *model.APIError)
	trendingForCurrentPeriod, trendingForLastPeriod := []model.Trending{}, []model.Trending{}
	today := time.Now()
	twoWeeksFromToday, fourWeeksFromToday := today.AddDate(0, 0, -14), today.AddDate(0, 0, -28)

	go getTrafficData(resource, twoWeeksFromToday, today, &trendingForCurrentPeriod, c1)
	go getTrafficData(resource, fourWeeksFromToday, twoWeeksFromToday, &trendingForLastPeriod, c2)

	if err1, err2 := <-c1, <-c2; err1 != nil {
		res.WriteHeader(err1.StatusCode)
		json.NewEncoder(res).Encode(err1)
	} else if err2 != nil {
		res.WriteHeader(err2.StatusCode)
		json.NewEncoder(res).Encode(err2)
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(trendingForCurrentPeriod)
}

func getTrafficData(r string, from time.Time, to time.Time, td *[]model.Trending, c chan *model.APIError) {
	var err *model.APIError
	*td, err = skcSuggestionEngineDBInterface.GetTrafficData(r, from, to)
	c <- err
}
