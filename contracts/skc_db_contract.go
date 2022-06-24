package contracts

type Card struct {
	CardID     string `db:"card_number" json:"cardID"`
	CardName   string `db:"card_name" json:"cardName"`
	CardEffect string `db:"card_effect" json:"cardEffect"`
}
