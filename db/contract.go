package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Card struct {
	CardID         string `db:"card_number" json:"cardID"`
	CardColor      string `db:"card_color" json:"cardColor"`
	CardName       string `db:"card_name" json:"cardName"`
	CardAttribute  string `db:"card_attribute" json:"cardAttribute"`
	CardEffect     string `db:"card_effect" json:"cardEffect"`
	MonsterType    string `db:"monster_type" json:"monsterType"`
	MonsterAttack  uint16 `db:"monster_attack" json:"monsterAttack"`
	MonsterDefense uint16 `db:"monster_defense" json:"monsterDefense"`
}

type DeckList struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	ListContent string             `bson:"contents"`
	VideoUrl    string             `bson:"videoUrl"`
	Tags        []string           `bson:"tags"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
