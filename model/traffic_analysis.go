package model

import (
	"time"

	"github.com/ygo-skc/skc-suggestion-engine/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrafficAnalysis struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Timestamp        time.Time          `bson:"timestamp" json:"timestamp"`
	Source           TrafficSource      `bson:"source" json:"source"`
	ResourceUtilized Resource           `bson:"resourceUtilized" json:"resourceUtilized"`
	UserData         UserData           `bson:"userData" json:"userData"`
}

type TrafficSource struct {
	SystemName string `bson:"systemName" json:"systemName" validate:"systemname"`
	Version    string `bson:"version" json:"version" validate:"systemversion"`
}

type UserData struct {
	IP       string   `bson:"ip" json:"ip"`
	Location Location `bson:"location" json:"location"`
}

type Location struct {
	City    string `bson:"city" json:"city"`
	Zip     string `bson:"zip" json:"zip"`
	Country string `bson:"country" json:"country"`
}

type Resource struct {
	Name  string `bson:"name" json:"name" validate:"required"`
	Value string `bson:"value" json:"value" validate:"required"`
}

type TrafficAnalysisInput struct {
	IP               string         `json:"ip" validate:"ipv4"`
	Source           *TrafficSource `json:"source" validate:"required"`
	ResourceUtilized *Resource      `json:"resourceUtilized" validate:"required"`
}

func (tai TrafficAnalysisInput) Validate() *APIError {
	if err := util.V.Struct(tai); err != nil {
		return &APIError{Message: util.HandleValidationErrors(err)}
	} else {
		return nil
	}
}
