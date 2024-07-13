package model

type CardOfTheDay struct {
	Date    string `bson:"date" json:"date"`
	Version uint8  `bson:"version" json:"version"`
	CardID  string `bson:"cardID" json:"-"`
	Card    Card   `bson:"-" json:"card"`
}
