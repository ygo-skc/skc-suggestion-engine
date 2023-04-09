package model

type CardOfTheDay struct {
	CardID  string `bson:"cardID" json:"cardID"`
	Date    string `bson:"date" json:"date"`
	Version uint8  `bson:"version" json:"version"`
}
