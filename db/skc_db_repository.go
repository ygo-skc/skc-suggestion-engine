package db

import (
	"log"
	"strings"
)

const (
	queryCardUsingCardID   string = "SELECT card_number, card_name, card_effect FROM cards WHERE card_number = ?"
	queryCardUsingCardName string = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_name = ?"
)

// Uses card ID to find instance of card.
// Returns error if no instance of card ID as found in DB.
func FindDesiredCardInDBUsingID(cardID string) (Card, error) {
	var card Card

	if err := skcDBConn.QueryRow(queryCardUsingCardID, cardID).Scan(&card.CardID, &card.CardName, &card.CardEffect); err != nil {
		return card, err
	}

	return card, nil
}

func FindDesiredCardInDBUsingMultipleCardIDs(cards []string) (map[string]Card, error) {
	args := make([]interface{}, len(cards))
	for index, cardId := range cards {
		args[index] = cardId
	}
	cardData := map[string]Card{}

	if rows, err := skcDBConn.Query("SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_number IN (?"+strings.Repeat(",?", len(args)-1)+")", args...); err != nil {
		log.Println("Error occurred while querying SKC DB for card info using 1 or more CardIDs", err)
		return nil, err
	} else {
		for rows.Next() {
			var card Card
			if err := rows.Scan(&card.CardID, &card.CardColor, &card.CardName, &card.CardAttribute, &card.CardEffect, &card.MonsterType, &card.MonsterAttack, &card.MonsterDefense); err != nil {
				log.Println("Error transforming row to Card object from SKC DB while using 1 or more CardIDs", err)
				return nil, err
			}

			cardData[card.CardID] = card
		}
	}

	return cardData, nil
}

// Uses card name to find instance of card.
// Returns error if no instance of card name as found in DB.
func FindDesiredCardInDBUsingName(cardName string) (Card, error) {
	var card Card

	if err := skcDBConn.QueryRow(queryCardUsingCardName, cardName).Scan(&card.CardID, &card.CardColor, &card.CardName, &card.CardAttribute, &card.CardEffect, &card.MonsterType, &card.MonsterAttack, &card.MonsterDefense); err != nil {
		return card, err
	}

	return card, nil
}
