package testing

import (
	"context"
	"log"
	"time"

	cModel "github.com/ygo-skc/skc-go/common/model"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

type SKCSuggestionEngineDAOImplementation struct{}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetSKCSuggestionDBVersion(ctx context.Context) (string, error) {
	return "1.0.0", nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) InsertTrafficData(ctx context.Context, ta model.TrafficAnalysis) *cModel.APIError {
	log.Fatalln("InsertTrafficData() not mocked")
	return nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetTrafficData(
	ctx context.Context, resourceName model.ResourceName, from time.Time, to time.Time) ([]model.TrafficResourceUtilizationMetric, *cModel.APIError) {
	log.Fatalln("GetTrafficData() not mocked")
	return nil, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) IsBlackListed(ctx context.Context, blackListType string, blackListPhrase string) (bool, *cModel.APIError) {
	log.Fatalln("IsBlackListed() not mocked")
	return false, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetCardOfTheDay(ctx context.Context, date string, version int) (*string, *cModel.APIError) {
	log.Fatalln("GetCardOfTheDay() not mocked")
	return nil, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) GetHistoricalCardOfTheDayData(ctx context.Context, version int) ([]string, *cModel.APIError) {
	log.Fatalln("GetHistoricalCardOfTheDayData() not mocked")
	return nil, nil
}

func (dbInterface SKCSuggestionEngineDAOImplementation) InsertCardOfTheDay(ctx context.Context, cotd model.CardOfTheDay) *cModel.APIError {
	log.Fatalln("InsertCardOfTheDay() not mocked")
	return nil
}
