package db

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ygo-skc/skc-suggestion-engine/model"
)

var (
	skcDBConn *sql.DB
)

const (
	// errors
	genericError string = "Error occurred while querying DB."

	// queries
	queryDBVersion                  string = "SELECT VERSION()"
	queryCardColorIDs               string = "select color_id, card_color from card_colors order by color_id"
	queryCardUsingCardID            string = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_number = ?"
	queryCardUsingCardName          string = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_name = ?"
	findRelatedCardsUsingCardEffect string = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE (card_effect LIKE ? OR card_effect LIKE ?) AND card_number != ? ORDER BY color_id, card_name"
	queryRandomCardID               string = "SELECT card_number FROM card_info WHERE card_color != 'Token' ORDER BY RAND() LIMIT 1"
)

// interface
type SKCDatabaseAccessObject interface {
	GetSKCDBVersion() (string, error)

	GetCardColorIDs() (map[string]int, *model.APIError)

	FindDesiredCardInDBUsingID(cardID string) (*model.Card, *model.APIError)
	FindDesiredCardInDBUsingMultipleCardIDs(cards []string) (model.CardDataMap, *model.APIError)
	FindDesiredCardInDBUsingName(cardName string) (model.Card, error)

	FindOccurrenceOfCardNameInAllCardEffect(cardName string, cardId string) ([]model.Card, *model.APIError)

	FindInArchetypeSupportUsingCardName(archetypeName string) ([]model.Card, *model.APIError)
	FindInArchetypeSupportUsingCardText(archetypeName string) ([]model.Card, *model.APIError)
	FindArchetypeExclusionsUsingCardText(archetypeName string) ([]model.Card, *model.APIError)

	FindDesiredProductInDBUsingMultipleProductIDs(cards []string) (model.ProductDataMap, *model.APIError)

	GetRandomCard() (string, *model.APIError)
}

// impl
type SKCDAOImplementation struct{}

// Get version of MYSQL being used by SKC DB.
func (imp SKCDAOImplementation) GetSKCDBVersion() (string, error) {
	var version string
	if err := skcDBConn.QueryRow(queryDBVersion).Scan(&version); err != nil {
		log.Println("Error getting SKC DB version", err)
		return version, err
	}

	return version, nil
}

// Get IDs for all card colors currently in database.
func (imp SKCDAOImplementation) GetCardColorIDs() (map[string]int, *model.APIError) {
	cardColorIDs := map[string]int{}

	if rows, err := skcDBConn.Query(queryCardColorIDs); err != nil {
		log.Printf("Error occurred while fetching card color IDs.")
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		for rows.Next() {
			var colorId int
			var cardColor string

			if err := rows.Scan(&colorId, &cardColor); err != nil {
				log.Println("Error reading row...", err)
				return cardColorIDs, &model.APIError{Message: "Error parsing data from DB."}
			}

			cardColorIDs[cardColor] = colorId
		}
	}
	return cardColorIDs, nil
}

// Uses card ID to find instance of card.
// Returns error if no instance of card ID is found in DB or other issues occur while accessing DB.
func (imp SKCDAOImplementation) FindDesiredCardInDBUsingID(cardID string) (*model.Card, *model.APIError) {
	var card model.Card

	if err := skcDBConn.QueryRow(queryCardUsingCardID, cardID).Scan(&card.CardID, &card.CardColor, &card.CardName, &card.CardAttribute, &card.CardEffect, &card.MonsterType, &card.MonsterAttack, &card.MonsterDefense); err != nil {
		if err.Error() == "sql: no rows in result set" {
			log.Printf("Card w/ ID {%s} not found in DB", cardID)
			return nil, &model.APIError{Message: fmt.Sprintf("Cannot find card using ID %s", cardID), StatusCode: http.StatusNotFound}
		} else {
			log.Printf("An error ocurred while fetching card using ID. Err {%s}", err)
			return nil, &model.APIError{Message: "Service unavailable", StatusCode: http.StatusInternalServerError}
		}
	}

	return &card, nil
}

func (imp SKCDAOImplementation) FindDesiredCardInDBUsingMultipleCardIDs(cardIDs []string) (model.CardDataMap, *model.APIError) {
	numCards := len(cardIDs)
	args := make([]interface{}, numCards)
	cardData := make(map[string]model.Card, numCards)

	for index, cardId := range cardIDs {
		args[index] = cardId
	}

	query := fmt.Sprintf("SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_number IN (%s)", variablePlaceholders(numCards))

	if rows, err := skcDBConn.Query(query, args...); err != nil {
		log.Println("Error occurred while querying SKC DB for card info using 1 or more CardIDs", err)
		return nil, &model.APIError{Message: genericError}
	} else {
		if cards, err := parseRowsForCard(rows); err != nil {
			return nil, err
		} else {
			for _, card := range cards {
				cardData[card.CardID] = card
			}
		}
	}

	return cardData, nil
}

func (imp SKCDAOImplementation) FindDesiredProductInDBUsingMultipleProductIDs(products []string) (model.ProductDataMap, *model.APIError) {
	numProducts := len(products)
	args := make([]interface{}, numProducts)
	productData := make(map[string]model.Product, numProducts)

	for index, cardId := range products {
		args[index] = cardId
	}

	query := fmt.Sprintf("SELECT product_id, product_locale, product_name, product_release_date, product_content_total, product_type, product_sub_type FROM product_info WHERE product_id IN (%s)", variablePlaceholders(numProducts))

	if rows, err := skcDBConn.Query(query, args...); err != nil {
		log.Println("Error occurred while querying SKC DB for product info using 1 or more Product IDs", err)
		return nil, &model.APIError{Message: genericError}
	} else {
		for rows.Next() {
			var product model.Product
			if err := rows.Scan(&product.ProductID, &product.ProductLocale,
				&product.ProductName, &product.ProductReleaseDate, &product.ProductTotal, &product.ProductType, &product.ProductSubType); err != nil {
				log.Println("Error transforming row to Product object from SKC DB while using 1 or more Product ID's", err)
				return nil, &model.APIError{Message: "Error parsing data from DB."}
			}

			productData[product.ProductID] = product
		}
	}

	return productData, nil
}

// Uses card name to find instance of card.
// Returns error if no instance of card name as found in DB.
func (imp SKCDAOImplementation) FindDesiredCardInDBUsingName(cardName string) (model.Card, error) {
	var card model.Card

	if err := skcDBConn.QueryRow(queryCardUsingCardName, cardName).
		Scan(&card.CardID, &card.CardColor, &card.CardName, &card.CardAttribute, &card.CardEffect, &card.MonsterType, &card.MonsterAttack, &card.MonsterDefense); err != nil {
		return card, err
	}

	return card, nil
}

// TODO: document
// TODO: find way to make code more readable
func (imp SKCDAOImplementation) FindOccurrenceOfCardNameInAllCardEffect(cardName string, cardId string) ([]model.Card, *model.APIError) {
	cardNameWithDoubleQuotes := `%"` + cardName + `"%`
	cardNameWithSingleQuotes := `%'` + cardName + `'%`

	if rows, err := skcDBConn.Query(findRelatedCardsUsingCardEffect, cardNameWithDoubleQuotes, cardNameWithSingleQuotes, cardId); err != nil {
		log.Printf("Error occurred while searching for occurrences of %s in all card effects. Err %v", cardName, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		return parseRowsForCard(rows)
	}
}

func (imp SKCDAOImplementation) FindInArchetypeSupportUsingCardName(archetypeName string) ([]model.Card, *model.APIError) {
	// there are three scenarios
	// - archetype reference could be in the beginning of the name
	search1 := archetypeName + ` %`
	// - archetype reference could be in the middle of the name
	search2 := `% ` + archetypeName + ` %`
	// - archetype reference could be in the end of the name
	search3 := `% ` + archetypeName

	if rows, err := skcDBConn.Query("SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_name LIKE ? OR card_name LIKE ? OR card_name LIKE ?", search1, search2, search3); err != nil {
		log.Printf("Error occurred while searching for in-archetype cards using archetype name %s. Err %v", archetypeName, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		return parseRowsForCard(rows)
	}
}

func (imp SKCDAOImplementation) FindInArchetypeSupportUsingCardText(archetypeName string) ([]model.Card, *model.APIError) {
	archetypeName = `%` + fmt.Sprintf(`This card is always treated as an "%s" card`, archetypeName) + `%`

	if rows, err := skcDBConn.Query("SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_effect LIKE ?", archetypeName); err != nil {
		log.Printf("Error occurred while searching for in-archetype cards using nickname %s. Err %v", archetypeName, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		return parseRowsForCard(rows)
	}
}

func (imp SKCDAOImplementation) FindArchetypeExclusionsUsingCardText(archetypeName string) ([]model.Card, *model.APIError) {
	archetypeName = `%` + fmt.Sprintf(`This card is not treated as a "%s" card`, archetypeName) + `%`

	if rows, err := skcDBConn.Query("SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_effect LIKE ?", archetypeName); err != nil {
		log.Printf("Error occurred while searching for in-archetype cards using nickname %s. Err %v", archetypeName, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		return parseRowsForCard(rows)
	}
}

func (imp SKCDAOImplementation) GetRandomCard() (string, *model.APIError) {
	var randomCardId string

	if err := skcDBConn.QueryRow(queryRandomCardID).Scan(&randomCardId); err != nil {
		log.Printf("Error occurred while fetching random card ID from database. Err %v", err)
		return "", &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	}
	return randomCardId, nil
}

func parseRowsForCard(rows *sql.Rows) ([]model.Card, *model.APIError) {
	cards := []model.Card{}

	for rows.Next() {
		var card model.Card
		if err := rows.Scan(&card.CardID, &card.CardColor, &card.CardName, &card.CardAttribute, &card.CardEffect, &card.MonsterType, &card.MonsterAttack, &card.MonsterDefense); err != nil {
			log.Printf("Error occurred while parsing results: %v.", err)
			return nil, &model.APIError{Message: "Error parsing card data from DB.", StatusCode: http.StatusInternalServerError}
		} else {
			cards = append(cards, card)
		}
	}

	return cards, nil // no parsing error
}

func variablePlaceholders(totalFields int) string {
	if totalFields == 0 {
		return ""
	} else if totalFields == 1 {
		return "?"
	} else {
		return fmt.Sprintf("?%s", strings.Repeat(", ?", totalFields-1))
	}
}
