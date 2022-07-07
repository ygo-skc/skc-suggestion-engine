package model

import "strings"

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

func (c Card) isExtraDeckMonster() bool {
	color := strings.ToUpper(c.CardEffect)
	return strings.Contains(color, "FUSION") || strings.Contains(color, "SYNCHRO") || strings.Contains(color, "XYZ") || strings.Contains(color, "PENDULUM") || strings.Contains(color, "LINK")
}
