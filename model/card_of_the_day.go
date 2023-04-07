package model

type CardOfTheDAy struct {
	CardID  string `json:"cardID"`
	Date    string `json:"date"`
	Version uint8  `json:"version"`
}
