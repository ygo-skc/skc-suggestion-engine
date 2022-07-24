package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

// Endpoint will allow clients to submit traffic data to be saved in a MongoDB instance.
func submitNewTrafficData(res http.ResponseWriter, req *http.Request) {
	log.Println("Adding new traffic record...")
	res.Header().Add("Content-Type", "application/json") // prepping res headers

	// verify client can call endpoint
	if err := verifyApiKey(req.Header); err != nil {
		res.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(res).Encode(err)
		return
	}

	// deserialize body
	var trafficData model.TrafficAnalysisInput
	if b, err := ioutil.ReadAll(req.Body); err != nil {
		log.Println("Error occurred while reading the request body.")

		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(model.APIError{Message: "Body could not be deserialize body."})
		return
	} else {
		json.Unmarshal(b, &trafficData)
	}

	// validate body
	if err := trafficData.Validate(); err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	}

	// get IP number info
	if ipData, err := ipDB.Get_all(trafficData.IP); err != nil {
		log.Printf("Error getting info for IP address %s. Error %v", trafficData.IP, err)

		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	} else {
		// create object to insert into collection
		location := model.Location{Zip: ipData.Zipcode, City: ipData.City, Country: ipData.Country_short}
		userData := model.UserData{Location: location, IP: trafficData.IP}
		source := model.TrafficSource{SystemName: trafficData.Source.SystemName, Version: trafficData.Source.Version}
		trafficAnalysis := model.TrafficAnalysis{Timestamp: time.Now(), UserData: userData, ResourceUtilized: *trafficData.ResourceUtilized, Source: source}

		db.InsertTrafficData(trafficAnalysis)

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(model.Success{Message: "Successfully inserted new traffic data."})
	}
}
