package model

type CardOfTheDay struct {
	CardID  string `json:"cardID"`
	Date    string `json:"date"`
	Version uint8  `json:"version"`
}
