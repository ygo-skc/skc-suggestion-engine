package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

// Endpoint will allow clients to submit traffic data to be saved in a MongoDB instance.
func submitNewTrafficDataHandler(res http.ResponseWriter, req *http.Request) {
	logger, ctx := cUtil.NewRequestSetup(context.Background(), "traffic data submission")
	logger.Info("Adding new traffic record")

	// deserialize body
	var trafficData model.TrafficData
	if err := json.NewDecoder(req.Body).Decode(&trafficData); err != nil {
		logger.Error("Error occurred while reading the request body")
		cModel.HandleServerResponse(cModel.APIError{Message: "Body could not be deserialized.", StatusCode: http.StatusBadRequest}, res)
		return
	}

	// validate body
	if err := validation.Validate(trafficData); err != nil {
		err.HandleServerResponse(res)
		return
	}

	// ensure resource is valid before storing it
	switch trafficData.ResourceUtilized.Name {
	case model.CardResource:
		if _, err := skcDBInterface.GetDesiredCardInDBUsingID(ctx, trafficData.ResourceUtilized.Value); err != nil {
			logger.Error(fmt.Sprintf("Card resource %s not valid", trafficData.ResourceUtilized.Value))
			res.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(res).Encode(cModel.APIError{Message: "Resource is not valid"})
			return
		}
	case model.ProductResource:
		if _, err := skcDBInterface.GetDesiredProductInDBUsingID(ctx, trafficData.ResourceUtilized.Value); err != nil {
			logger.Error(fmt.Sprintf("Product resource %s not valid", trafficData.ResourceUtilized.Value))
			res.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(res).Encode(cModel.APIError{Message: "Resource is not valid"})
			return
		}
	}

	// get IP number info
	var location model.Location
	if ipData, err := ipDB.Get_all(trafficData.IP); err != nil {
		logger.Error(fmt.Sprintf("Error getting info for IP address %s. Error %v", trafficData.IP, err))

		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(cModel.APIError{Message: "The IP provided was not found in the IP Database. Therefor, not storing traffic pattern."})
		return
	} else {
		location = model.Location{Zip: ipData.Zipcode, City: ipData.City, Country: ipData.Country_short}
	}

	// create traffic analysis object that will be inserted to DB
	userData := model.UserData{Location: location, IP: trafficData.IP}
	source := model.TrafficSource{SystemName: trafficData.Source.SystemName, Version: trafficData.Source.Version}
	trafficAnalysis := model.TrafficAnalysis{Timestamp: time.Now(), UserData: userData, ResourceUtilized: *trafficData.ResourceUtilized, Source: source}

	if err := skcSuggestionEngineDBInterface.InsertTrafficData(ctx, trafficAnalysis); err != nil {
		err.HandleServerResponse(res)
		return
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(cModel.Success{Message: "Successfully inserted new traffic data."})
}

func trending(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	r := model.ResourceName(strings.ToUpper(pathVars["resource"]))

	logger, ctx := cUtil.NewRequestSetup(context.Background(), "trending", slog.String("resource", string(r)))
	logger.Info("Getting trending data")

	c1, c2 := make(chan *cModel.APIError), make(chan *cModel.APIError)
	metricsForCurrentPeriod, metricsForLastPeriod := []model.TrafficResourceUtilizationMetric{}, []model.TrafficResourceUtilizationMetric{}
	today := time.Now()
	firstInterval, secondInterval := today.AddDate(0, 0, -10), today.AddDate(0, 0, -20)

	go getMetrics(ctx, r, firstInterval, today, &metricsForCurrentPeriod, c1)
	go getMetrics(ctx, r, secondInterval, firstInterval, &metricsForLastPeriod, c2)

	// verify go routines exited with no errors
	for i := 0; i < 2; i++ {
		select {
		case err := <-c1:
			if err != nil {
				err.HandleServerResponse(res)
				return
			}
		case err := <-c2:
			if err != nil {
				err.HandleServerResponse(res)
				return
			}
		}
	}

	if c3, afterResourcesAreFetchedCB := initResourceInfoFlow(ctx, r, metricsForCurrentPeriod); c3 == nil || afterResourcesAreFetchedCB == nil {
		(&cModel.APIError{StatusCode: 500, Message: "Using incorrect resource name."}).HandleServerResponse(res)
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

func initResourceInfoFlow(ctx context.Context, r model.ResourceName, metricsForCurrentPeriod []model.TrafficResourceUtilizationMetric) (chan *cModel.APIError, func([]model.TrendingMetric)) {
	c := make(chan *cModel.APIError)

	switch r {
	case model.CardResource:
		cdm := &cModel.BatchCardData[cModel.CardIDs]{}
		go fetchResourceInfo[cModel.CardIDs](ctx, metricsForCurrentPeriod, cdm, skcDBInterface.GetDesiredCardInDBUsingMultipleCardIDs, c)
		return c, func(tm []model.TrendingMetric) { updateTrendingMetric(tm, metricsForCurrentPeriod, cdm.CardInfo) }
	case model.ProductResource:
		pdm := &cModel.BatchProductData[cModel.ProductIDs]{}
		go fetchResourceInfo[cModel.ProductIDs](ctx, metricsForCurrentPeriod, pdm, skcDBInterface.GetDesiredProductInDBUsingMultipleProductIDs, c)
		return c, func(tm []model.TrendingMetric) {
			updateTrendingMetric(tm, metricsForCurrentPeriod, pdm.ProductInfo)
		}
	}
	return nil, nil
}

func updateTrendingMetric[T cModel.Card | cModel.Product](tm []model.TrendingMetric, metricsForCurrentPeriod []model.TrafficResourceUtilizationMetric, dataMap map[string]T) {
	for ind := range tm {
		tm[ind].Resource = dataMap[metricsForCurrentPeriod[ind].ResourceValue]
	}
}

func fetchResourceInfo[IS cModel.IdentifierSlice, BD cModel.BatchData[IS]](ctx context.Context,
	metrics []model.TrafficResourceUtilizationMetric, bathData *BD, fetchResourceFromDB func(context.Context, []string) (BD, *cModel.APIError), c chan<- *cModel.APIError) {
	rv := make([]string, len(metrics))
	for ind, value := range metrics {
		rv[ind] = value.ResourceValue
	}

	if bri, err := fetchResourceFromDB(ctx, rv); err != nil {
		cUtil.LoggerFromContext(ctx).Info("Could not fetch data for trending resources")
		c <- err
	} else {
		*bathData = bri
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

func getMetrics(ctx context.Context, r model.ResourceName, from time.Time, to time.Time, td *[]model.TrafficResourceUtilizationMetric, c chan<- *cModel.APIError) {
	var err *cModel.APIError
	if *td, err = skcSuggestionEngineDBInterface.GetTrafficData(ctx, r, from, to); err != nil {
		cUtil.LoggerFromContext(ctx).Error(fmt.Sprintf("There was an issue fetching traffic data for starting date %v and ending date %v", from, to))
	}
	c <- err
}
