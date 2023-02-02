package model

type CardSupport struct {
	Card    *Card   `json:"card"`
	Support *[]Card `json:"support"`
}
