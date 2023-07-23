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
	var trafficData model.TrafficData
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
	metricsForCurrentPeriod, metricsForLastPeriod := []model.TrafficResourceUtilizationMetric{}, []model.TrafficResourceUtilizationMetric{}
	today := time.Now()
	twoWeeksFromToday, fourWeeksFromToday := today.AddDate(0, 0, -14), today.AddDate(0, 0, -28)

	go getMetrics(resource, twoWeeksFromToday, today, &metricsForCurrentPeriod, c1)
	go getMetrics(resource, fourWeeksFromToday, twoWeeksFromToday, &metricsForLastPeriod, c2)

	// get channel data and check for errors
	if err1, err2 := <-c1, <-c2; err1 != nil {
		res.WriteHeader(err1.StatusCode)
		json.NewEncoder(res).Encode(err1)
	} else if err2 != nil {
		res.WriteHeader(err2.StatusCode)
		json.NewEncoder(res).Encode(err2)
	}

	c3 := make(chan *model.APIError)
	cdm := model.CardDataMap{}
	go fetchResourceInfo(metricsForCurrentPeriod, &cdm, c3)

	tm := determineTrendChange(metricsForCurrentPeriod, metricsForLastPeriod)

	if err1 := <-c3; err1 != nil {
		res.WriteHeader(err1.StatusCode)
		json.NewEncoder(res).Encode(err1)
	}

	for ind := range tm {
		tm[ind].Resource = cdm[metricsForCurrentPeriod[ind].ResourceValue]
	}

	trending := model.Trending[model.Card]{ResourceName: resource, Metrics: tm}
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(trending)
}

func fetchResourceInfo(metrics []model.TrafficResourceUtilizationMetric, cdm *model.CardDataMap, c chan *model.APIError) {
	rv := make([]string, len(metrics))

	for ind, value := range metrics {
		rv[ind] = value.ResourceValue
	}

	if resourceData, err := skcDBInterface.FindDesiredCardInDBUsingMultipleCardIDs(rv); err == nil {
		for k, v := range resourceData {
			(*cdm)[k] = v
		}
		c <- nil
	} else {
		log.Printf("Could not fetch card info for trending data...")
		c <- err
	}
}

func determineTrendChange[R model.TrafficResourceType](metricsForCurrentPeriod []model.TrafficResourceUtilizationMetric, metricsForLastPeriod []model.TrafficResourceUtilizationMetric) []model.TrendingMetric[R] {
	totalElements := len(metricsForCurrentPeriod)
	previousPeriodRanking := make(map[string]int, totalElements)
	tm := make([]model.TrendingMetric[R], totalElements)

	for ind, value := range metricsForLastPeriod {
		previousPeriodRanking[value.ResourceValue] = ind
	}

	for currentPeriodPosition, value := range metricsForCurrentPeriod {
		if previousPeriodPosition, isPresent := previousPeriodRanking[value.ResourceValue]; isPresent {
			tm[currentPeriodPosition] = model.TrendingMetric[R]{
				Change:      previousPeriodPosition - currentPeriodPosition,
				Occurrences: value.Occurrences,
			}
		} else {
			tm[currentPeriodPosition] = model.TrendingMetric[R]{
				Change:      totalElements - currentPeriodPosition,
				Occurrences: value.Occurrences,
			}
		}
	}

	return tm
}

func getMetrics(r string, from time.Time, to time.Time, td *[]model.TrafficResourceUtilizationMetric, c chan *model.APIError) {
	var err *model.APIError
	*td, err = skcSuggestionEngineDBInterface.GetTrafficData(r, from, to)
	c <- err
}
