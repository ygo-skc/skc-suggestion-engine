package model

import (
	cModel "github.com/ygo-skc/skc-go/common/model"
)

type CardOfTheDay struct {
	Date    string      `bson:"date" json:"date"`
	Version int         `bson:"version" json:"version"`
	CardID  string      `bson:"cardID" json:"-"`
	Card    cModel.Card `bson:"-" json:"card"`
}
