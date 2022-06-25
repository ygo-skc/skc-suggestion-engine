package db

const (
	queryCardUsingCardID   string = "SELECT card_number, card_name, card_effect FROM cards WHERE card_number = ?"
	queryCardUsingCardName string = "SELECT card_number, card_name, card_effect FROM cards WHERE card_name = ?"
)

func FindDesiredCardInDBUsingID(cardID string) (Card, error) {
	var card Card

	if err := skcDBConn.QueryRow(queryCardUsingCardID, cardID).Scan(&card.CardID, &card.CardName, &card.CardEffect); err != nil {
		return card, err
	}

	return card, nil
}

func FindDesiredCardInDBUsingName(cardName string) (Card, error) {
	var card Card
	if err := skcDBConn.QueryRow(queryCardUsingCardName, cardName).Scan(&card.CardID, &card.CardName, &card.CardEffect); err != nil {
		return card, err
	}

	return card, nil
}
