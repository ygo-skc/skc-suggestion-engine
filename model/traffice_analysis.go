package model

import (
	"time"

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
	Name string `bson:"name" json:"name"`
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
	Name  string `bson:"name" json:"name"`
	Value string `bson:"value" json:"value"`
}

type TrafficAnalysisInput struct {
	IP               string   `json:"ip" validate:"ipv4"`
	SourceName       string   `json:"sourceName"`
	ResourceUtilized Resource `json:"resourceUtilized"`
}
