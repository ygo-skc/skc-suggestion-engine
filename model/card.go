package model

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
