package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

// Endpoint will allow clients to submit traffic data to be saved in a MongoDB instance.
func submitNewTrafficDataHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Adding new traffic record...")

	// deserialize body
	var trafficData model.TrafficData
	if err := json.NewDecoder(req.Body).Decode(&trafficData); err != nil {
		log.Println("Error occurred while reading the request body.")
		model.HandleServerResponse(model.APIError{Message: "Body could not be deserialized.", StatusCode: http.StatusBadRequest}, res)
		return
	}

	// validate body
	if err := validation.Validate(trafficData); err != nil {
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
		err.HandleServerResponse(res)
		return
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(model.Success{Message: "Successfully inserted new traffic data."})
}

func trending(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	r := model.ResourceName(strings.ToUpper(pathVars["resource"]))
	log.Printf("Getting trending data for resource: %s", r)

	c1, c2 := make(chan *model.APIError), make(chan *model.APIError)
	metricsForCurrentPeriod, metricsForLastPeriod := []model.TrafficResourceUtilizationMetric{}, []model.TrafficResourceUtilizationMetric{}
	today := time.Now()
	twoWeeksFromToday, fourWeeksFromToday := today.AddDate(0, 0, -14), today.AddDate(0, 0, -28)

	go getMetrics(r, twoWeeksFromToday, today, &metricsForCurrentPeriod, c1)
	go getMetrics(r, fourWeeksFromToday, twoWeeksFromToday, &metricsForLastPeriod, c2)

	// get channel data and check for errors
	if err1, err2 := <-c1, <-c2; err1 != nil {
		err1.HandleServerResponse(res)
		return
	} else if err2 != nil {
		err2.HandleServerResponse(res)
		return
	}

	if c3, afterResourcesAreFetchedCB := initResourceInfoFlow(r, metricsForCurrentPeriod); c3 == nil || afterResourcesAreFetchedCB == nil {
		(&model.APIError{StatusCode: 500, Message: "Using incorrect resource name."}).HandleServerResponse(res)
		return
	} else {
		tm := determineTrendChange(metricsForCurrentPeriod, metricsForLastPeriod)
		trending := model.Trending{ResourceName: r, Metrics: tm}

		if err1 := <-c3; err1 != nil {
			err1.HandleServerResponse(res)
			return
		}

		afterResourcesAreFetchedCB(tm)
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(trending)
	}
}

func initResourceInfoFlow(r model.ResourceName, metricsForCurrentPeriod []model.TrafficResourceUtilizationMetric) (chan *model.APIError, func([]model.TrendingMetric)) {
	c := make(chan *model.APIError)

	switch r {
	case model.CardResource:
		cdm := &model.BatchCardInfo{}
		go fetchResourceInfo(metricsForCurrentPeriod, cdm, skcDBInterface.GetDesiredCardInDBUsingMultipleCardIDs, c)
		return c, func(tm []model.TrendingMetric) { updateTrendingMetric(tm, metricsForCurrentPeriod, cdm.CardInfo) }
	case model.ProductResource:
		pdm := &model.BatchProductInfo{}
		go fetchResourceInfo(metricsForCurrentPeriod, pdm, skcDBInterface.GetDesiredProductInDBUsingMultipleProductIDs, c)
		return c, func(tm []model.TrendingMetric) { updateTrendingMetric(tm, metricsForCurrentPeriod, pdm.ProductInfo) }
	}
	return nil, nil
}

func updateTrendingMetric[T model.Card | model.Product](
	tm []model.TrendingMetric, metricsForCurrentPeriod []model.TrafficResourceUtilizationMetric, dataMap map[string]T) {
	for ind := range tm {
		tm[ind].Resource = dataMap[metricsForCurrentPeriod[ind].ResourceValue]
	}
}

func fetchResourceInfo[BD model.BatchData](
	metrics []model.TrafficResourceUtilizationMetric,
	bathData *BD,
	fetchResourceFromDB func([]string) (*BD, *model.APIError),
	c chan *model.APIError) {
	rv := make([]string, len(metrics))
	for ind, value := range metrics {
		rv[ind] = value.ResourceValue
	}

	if bri, err := fetchResourceFromDB(rv); err != nil {
		log.Printf("Could not fetch data for trending resources")
		c <- err
	} else {
		*bathData = *bri
	}

	c <- nil
}

func determineTrendChange(
	metricsForCurrentPeriod []model.TrafficResourceUtilizationMetric,
	metricsForLastPeriod []model.TrafficResourceUtilizationMetric,
) []model.TrendingMetric {
	totalElements := len(metricsForCurrentPeriod)
	previousPeriodRanking := make(map[string]int, totalElements)
	tm := make([]model.TrendingMetric, totalElements)

	for ind, value := range metricsForLastPeriod {
		previousPeriodRanking[value.ResourceValue] = ind
	}

	for currentPeriodPosition, value := range metricsForCurrentPeriod {
		tm[currentPeriodPosition] = model.TrendingMetric{Occurrences: value.Occurrences}

		if previousPeriodPosition, isPresent := previousPeriodRanking[value.ResourceValue]; isPresent {
			tm[currentPeriodPosition].Change = previousPeriodPosition - currentPeriodPosition
		} else {
			tm[currentPeriodPosition].Change = totalElements - currentPeriodPosition
		}
	}

	return tm
}

func getMetrics(r model.ResourceName, from time.Time, to time.Time, td *[]model.TrafficResourceUtilizationMetric, c chan *model.APIError) {
	var err *model.APIError
	if *td, err = skcSuggestionEngineDBInterface.GetTrafficData(r, from, to); err != nil {
		log.Printf("There was an issue fetching traffic data for starting date %v and ending date %v", from, to)
	}
	c <- err
}
