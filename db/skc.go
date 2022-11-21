package db

import (
	"fmt"
	"log"
	"strings"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	// queries
	queryDBVersion         string = "SELECT VERSION()"
	queryCardUsingCardID   string = "SELECT card_number, card_color, card_name, card_effect FROM card_info WHERE card_number = ?"
	queryCardUsingCardName string = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_name = ?"
)

// interface
type SKCDatabaseAccessObject interface {
	GetSKCDBVersion() (string, error)
	FindDesiredCardInDBUsingID(cardID string) (*model.Card, *model.APIError)
	FindDesiredCardInDBUsingMultipleCardIDs(cards []string) (model.DeckListContents, model.APIError)
	FindDesiredCardInDBUsingName(cardName string) (model.Card, error)
}

// impl
type SKCDatabaseAccessObjectImplementation struct{}

// Get version of MYSQL being used by SKC DB.
func (imp SKCDatabaseAccessObjectImplementation) GetSKCDBVersion() (string, error) {
	var version string
	if err := skcDBConn.QueryRow(queryDBVersion).Scan(&version); err != nil {
		log.Println("Error getting SKC DB version", err)
		return version, err
	}

	return version, nil
}

// Uses card ID to find instance of card.
// Returns error if no instance of card ID is found in DB or other issues occur while accessing DB.
func (imp SKCDatabaseAccessObjectImplementation) FindDesiredCardInDBUsingID(cardID string) (*model.Card, *model.APIError) {
	var card model.Card

	if err := skcDBConn.QueryRow(queryCardUsingCardID, cardID).Scan(&card.CardID, &card.CardColor, &card.CardName, &card.CardEffect); err != nil {
		if err.Error() == "sql: no rows in result set" {
			log.Printf("Card w/ ID {%s} not found in DB", cardID)
			return nil, &model.APIError{Message: fmt.Sprintf("Cannot find card using ID %s", cardID)}
		} else {
			log.Printf("An error ocurred while fetching card using ID. Err {%s}", err)
			return nil, &model.APIError{Message: "Service unavailable"}
		}
	}

	return &card, nil
}

func (imp SKCDatabaseAccessObjectImplementation) FindDesiredCardInDBUsingMultipleCardIDs(cards []string) (model.DeckListContents, model.APIError) {
	args := make([]interface{}, len(cards))
	for index, cardId := range cards {
		args[index] = cardId
	}
	cardData := map[string]model.Card{}

	if rows, err := skcDBConn.Query("SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_number IN (?"+strings.Repeat(",?", len(args)-1)+")", args...); err != nil {
		log.Println("Error occurred while querying SKC DB for card info using 1 or more CardIDs", err)
		return nil, model.APIError{Message: "Database could not be reached."}
	} else {
		for rows.Next() {
			var card model.Card
			if err := rows.Scan(&card.CardID, &card.CardColor, &card.CardName, &card.CardAttribute, &card.CardEffect, &card.MonsterType, &card.MonsterAttack, &card.MonsterDefense); err != nil {
				log.Println("Error transforming row to Card object from SKC DB while using 1 or more CardIDs", err)
				return nil, model.APIError{Message: "Error parsing data from DB."}
			}

			cardData[card.CardID] = card
		}
	}

	return cardData, model.APIError{}
}

// Uses card name to find instance of card.
// Returns error if no instance of card name as found in DB.
func (imp SKCDatabaseAccessObjectImplementation) FindDesiredCardInDBUsingName(cardName string) (model.Card, error) {
	var card model.Card

	if err := skcDBConn.QueryRow(queryCardUsingCardName, cardName).Scan(&card.CardID, &card.CardColor, &card.CardName, &card.CardAttribute, &card.CardEffect, &card.MonsterType, &card.MonsterAttack, &card.MonsterDefense); err != nil {
		return card, err
	}

	return card, nil
}
