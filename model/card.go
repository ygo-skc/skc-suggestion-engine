package model

import (
	"log"
	"strings"
)

type Card struct {
	CardID         string  `db:"card_number" json:"cardID"`
	CardColor      string  `db:"card_color" json:"cardColor"`
	CardName       string  `db:"card_name" json:"cardName"`
	CardAttribute  string  `db:"card_attribute" json:"cardAttribute"`
	CardEffect     string  `db:"card_effect" json:"cardEffect"`
	MonsterType    *string `db:"monster_type" json:"monsterType"`
	MonsterAttack  *uint16 `db:"monster_attack" json:"monsterAttack"`
	MonsterDefense *uint16 `db:"monster_defense" json:"monsterDefense"`
}

func (c Card) IsExtraDeckMonster() bool {
	color := strings.ToUpper(c.CardColor)
	return strings.Contains(color, "FUSION") || strings.Contains(color, "SYNCHRO") || strings.Contains(color, "XYZ") || strings.Contains(color, "PENDULUM") || strings.Contains(color, "LINK")
}

// Uses new line as delimiter to split card effect. Materials are found in the first token.
func (card Card) GetPotentialMaterialsAsString() string {
	var effectTokens []string
	if card.CardColor == "Pendulum-Fusion" {
		effectTokens = strings.SplitAfter(strings.SplitAfter(card.CardEffect, "\n\nMonster Effect\n")[1], "\n")
	} else {
		effectTokens = strings.SplitAfter(card.CardEffect, "\n")
	}

	if len(effectTokens) < 2 {
		log.Printf("Card w/ ID {%s} doesn't seem to have a materials string", card.CardID)
		return ""
	}

	return effectTokens[0]
}
