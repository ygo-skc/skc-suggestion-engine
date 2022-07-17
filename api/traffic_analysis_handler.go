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

func submitNewTrafficData(res http.ResponseWriter, req *http.Request) {
	log.Println("Adding new traffic record...")
	res.Header().Add("Content-Type", "application/json") // prepping res headers

	if err := verifyApiKey(req.Header); err.Message != "" {
		res.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(res).Encode(err)
		return
	}

	var body model.TrafficAnalysisInput

	if b, err := ioutil.ReadAll(req.Body); err != nil {
		log.Println("Error occurred while reading the request body.")

		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(model.APIError{Message: "Body could not be deserialize body."})
		return
	} else {
		json.Unmarshal(b, &body)
	}

	ipData, _ := ipDB.Get_all(body.IP)

	location := model.Location{Zip: ipData.Zipcode, City: ipData.City, Country: ipData.Country_short}
	userData := model.UserData{Location: location, IP: body.IP}
	trafficAnalysis := model.TrafficAnalysis{Timestamp: time.Now(), UserData: userData, ResourceUtilized: body.ResourceUtilized}

	db.XXX(trafficAnalysis)

	res.WriteHeader(http.StatusOK)
}
