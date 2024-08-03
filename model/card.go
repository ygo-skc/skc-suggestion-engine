package model

import (
	"strings"
)

type Card struct {
	CardID         string  `db:"card_number" json:"cardID"`
	CardColor      string  `db:"card_color" json:"cardColor"`
	CardName       string  `db:"card_name" json:"cardName"`
	CardAttribute  string  `db:"card_attribute" json:"cardAttribute"`
	CardEffect     string  `db:"card_effect" json:"cardEffect"`
	MonsterType    *string `db:"monster_type" json:"monsterType,omitempty"`
	MonsterAttack  *uint16 `db:"monster_attack" json:"monsterAttack,omitempty"`
	MonsterDefense *uint16 `db:"monster_defense" json:"monsterDefense,omitempty"`
}

func (c Card) IsExtraDeckMonster() bool {
	color := strings.ToUpper(c.CardColor)
	return strings.Contains(color, "FUSION") || strings.Contains(color, "SYNCHRO") || strings.Contains(color, "XYZ") || strings.Contains(color, "PENDULUM") || strings.Contains(color, "LINK")
}

// Uses new line as delimiter to split card effect. Materials are found in the first token.
func (card Card) GetPotentialMaterialsAsString() string {
	var effectTokens []string

	if !card.IsExtraDeckMonster() {
		return ""
	}

	color := strings.ToUpper(card.CardColor)
	if strings.Contains(color, "PENDULUM") && color != "PENDULUM-EFFECT" && color != "PENDULUM-NORMAL" {
		effectTokens = strings.SplitAfter(strings.SplitAfter(card.CardEffect, "\n\nMonster Effect\n")[1], "\n")
	} else {
		effectTokens = strings.SplitAfter(card.CardEffect, "\n")
	}

	if len(effectTokens) < 2 {
		return card.CardEffect
	}
	return effectTokens[0]
}

func (c Card) IsCardNameInTokens(tokens []QuotedToken) bool {
	isFound := false

	for _, token := range tokens {
		CleanupToken(&token)

		if strings.EqualFold(c.CardName, token) {
			isFound = true
			break
		}
	}

	return isFound
}

// cleans up a quoted string found in card text so its easier to parse
func CleanupToken(t *QuotedToken) {
	*t = strings.TrimSpace(*t)
	*t = strings.ReplaceAll(*t, `".`, "")
	*t = strings.ReplaceAll(*t, `",`, "")
	*t = strings.ReplaceAll(*t, "'.", "")
	*t = strings.ReplaceAll(*t, "',", "")

	*t = strings.Trim(*t, "'")
	*t = strings.Trim(*t, `"`)
}
