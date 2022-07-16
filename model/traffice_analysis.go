package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrafficAnalysis struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	Timestamp         time.Time          `bson:"timestamp" json:"timestamp"`
	Source            TrafficSource      `bson:"source" json:"source"`
	ResourceRequested string             `bson:"resourceRequested" json:"resourceRequested"`
	UserData          UserData           `bson:"userData" json:"userData"`
}

type TrafficSource struct {
	Name string `bson:"name" json:"name"`
}

type UserData struct {
	IP       string   `bson:"ip" json:"ip"`
	Location Location `bson:"location" json:"location"`
}

type Location struct {
}
