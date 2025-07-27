package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	json "github.com/goccy/go-json"
	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

const (
	trafficDataSubmissionOp = "Traffic Data Submission"
	trendingDataOp          = "Trending Data"
)

// Endpoint will allow clients to submit traffic data to be saved in a MongoDB instance.
func submitNewTrafficDataHandler(res http.ResponseWriter, req *http.Request) {
	logger, ctx := cUtil.InitRequest(context.Background(), apiName, trafficDataSubmissionOp)
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
		if _, err := downstream.YGO.CardService.GetCardByID(ctx, trafficData.ResourceUtilized.Value); err != nil {
			logger.Error(fmt.Sprintf("Card resource %s not valid", trafficData.ResourceUtilized.Value))
			res.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(res).Encode(cModel.APIError{Message: "Resource is not valid"})
			return
		}
	case model.ProductResource:
		if _, err := downstream.YGO.ProductService.GetProductSummaryByIDProto(ctx, trafficData.ResourceUtilized.Value); err != nil {
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
	resourceName := model.ResourceName(chi.URLParam(req, "resource"))

	logger, ctx := cUtil.InitRequest(context.Background(), apiName, trendingDataOp, slog.String("resource", string(resourceName)))
	logger.Info("Getting trending data")

	metricsForCurrentPeriod, metricsForLastPeriod := []model.TrafficResourceUtilizationMetric{}, []model.TrafficResourceUtilizationMetric{}
	today := time.Now()
	firstInterval, secondInterval := today.AddDate(0, 0, -10), today.AddDate(0, 0, -20)

	var wg sync.WaitGroup
	wg.Add(2)
	var e1, e2 atomic.Pointer[cModel.APIError]
	go getMetrics(ctx, resourceName, firstInterval, today, &metricsForCurrentPeriod, &e1, &wg)
	go getMetrics(ctx, resourceName, secondInterval, firstInterval, &metricsForLastPeriod, &e2, &wg)

	// verify go routines exited with no errors
	wg.Wait()
	if err := e1.Load(); err != nil {
		err.HandleServerResponse(res)
		return
	}
	if err := e2.Load(); err != nil {
		err.HandleServerResponse(res)
		return
	}

	if c3, addResourceInfoToTrendingMetric := fetchResourceInfoAsync(ctx, resourceName, metricsForCurrentPeriod); c3 == nil || addResourceInfoToTrendingMetric == nil {
		(&cModel.APIError{StatusCode: 500, Message: "Using incorrect resource name."}).HandleServerResponse(res)
		return
	} else {
		tm := determineTrendChange(metricsForCurrentPeriod, metricsForLastPeriod)
		trending := model.Trending{ResourceName: resourceName, Metrics: tm}

		if err1 := <-c3; err1 != nil {
			err1.HandleServerResponse(res)
			return
		}

		addResourceInfoToTrendingMetric(tm)
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(trending)
	}
}

func fetchResourceInfoAsync(ctx context.Context, r model.ResourceName, metricsForCurrentPeriod []model.TrafficResourceUtilizationMetric) (chan *cModel.APIError, func([]model.TrendingMetric)) {
	c := make(chan *cModel.APIError)

	switch r {
	case model.CardResource:
		cdm := &cModel.BatchCardData[cModel.CardIDs]{}
		go fetchResourceInfo(ctx, metricsForCurrentPeriod, &cdm, downstream.YGO.CardService.GetCardsByID, c)
		return c, func(tm []model.TrendingMetric) {
			updateTrendingMetric(tm, metricsForCurrentPeriod, cdm.CardInfo)
		}
	case model.ProductResource:
		pdm := &cModel.BatchProductSummaryData[cModel.ProductIDs]{}
		go fetchResourceInfo(ctx, metricsForCurrentPeriod, &pdm, downstream.YGO.ProductService.GetProductsSummaryByID, c)
		return c, func(tm []model.TrendingMetric) {
			updateTrendingMetric(tm, metricsForCurrentPeriod, pdm.ProductInfo)
		}
	}
	return nil, nil
}

func updateTrendingMetric[T cModel.YGOResource](tm []model.TrendingMetric, metricsForCurrentPeriod []model.TrafficResourceUtilizationMetric, dataMap map[string]T) {
	for ind := range tm {
		tm[ind].Resource = dataMap[metricsForCurrentPeriod[ind].ResourceValue]
	}
}

func fetchResourceInfo[RK cModel.YGOResourceKey, BD cModel.BatchCardData[RK] | cModel.BatchProductSummaryData[RK]](ctx context.Context,
	metrics []model.TrafficResourceUtilizationMetric, batchData **BD,
	fetchResourceFromDB func(context.Context, RK) (*BD, *cModel.APIError), c chan<- *cModel.APIError) {
	rv := make(RK, len(metrics))
	for ind, value := range metrics {
		rv[ind] = value.ResourceValue
	}

	if bri, err := fetchResourceFromDB(ctx, rv); err != nil {
		cUtil.RetrieveLogger(ctx).Info("Could not fetch data for trending resources")
		c <- err
	} else {
		*batchData = bri
	}

	c <- nil
}

func determineTrendChange(metricsForCurrentPeriod []model.TrafficResourceUtilizationMetric,
	metricsForLastPeriod []model.TrafficResourceUtilizationMetric) []model.TrendingMetric {
	totalElements := len(metricsForCurrentPeriod)
	previousPeriodPosition := make(map[string]int, totalElements)
	tm := make([]model.TrendingMetric, totalElements)

	for ind, value := range metricsForLastPeriod {
		previousPeriodPosition[value.ResourceValue] = ind
	}

	for currentPeriodPosition, value := range metricsForCurrentPeriod {
		tm[currentPeriodPosition] = model.TrendingMetric{Occurrences: value.Occurrences}

		if previousPeriodPosition, isPresent := previousPeriodPosition[value.ResourceValue]; isPresent {
			tm[currentPeriodPosition].Change = previousPeriodPosition - currentPeriodPosition
		} else {
			tm[currentPeriodPosition].Change = totalElements - currentPeriodPosition
		}
	}

	return tm
}

func getMetrics(ctx context.Context, r model.ResourceName, from time.Time, to time.Time,
	td *[]model.TrafficResourceUtilizationMetric, e *atomic.Pointer[cModel.APIError], wg *sync.WaitGroup) {
	defer wg.Done()
	var err *cModel.APIError
	*td, err = skcSuggestionEngineDBInterface.GetTrafficData(ctx, r, from, to)
	e.Store(err)
}
