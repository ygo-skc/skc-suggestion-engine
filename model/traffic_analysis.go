package model

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrafficAnalysis struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Timestamp        time.Time          `bson:"timestamp" json:"timestamp"`
	Source           TrafficSource      `bson:"source" json:"source"`
	ResourceUtilized TrafficResource    `bson:"resourceUtilized" json:"resourceUtilized"`
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

type TrafficResource struct {
	Name  string `bson:"name" json:"name" validate:"required"`
	Value string `bson:"value" json:"value" validate:"required"`
}

type TrafficData struct {
	IP               string           `json:"ip" validate:"ipv4"`
	Source           *TrafficSource   `json:"source" validate:"required"`
	ResourceUtilized *TrafficResource `json:"resourceUtilized" validate:"required"`
}

type TrafficResourceUtilizationMetric struct {
	ResourceValue string `bson:"_id" json:"resourceValue"`
	Occurrences   int    `json:"occurrences"`
}

type TrafficResourceType interface {
	Card
}

type Trending struct {
	ResourceName string           `json:"resourceName"`
	Metrics      []TrendingMetric `json:"metrics"`
}

type TrendingMetric struct {
	Resource    Card `json:"resource"`
	Occurrences int  `json:"occurrences"`
	Change      int  `json:"change"`
}

func (tai TrafficData) Validate() *validation.ValidationErrors {
	if err := validation.V.Struct(tai); err != nil {
		return validation.HandleValidationErrors(err.(validator.ValidationErrors))
	} else {
		return nil
	}
}
