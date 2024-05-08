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
	genericError = "Error occurred while querying DB"

	// logs
	queryErrorLog = "Error fetching data from DB - %v"
	parseErrorLog = "Error parsing data from DB - %v"

	// queries
	queryDBVersion    = "SELECT VERSION()"
	queryCardColorIDs = "SELECT color_id, card_color from card_colors ORDER BY color_id"

	queryCardUsingCardID     = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_number = ?"
	queryCardUsingCardIDs    = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_number IN (%s)"
	queryCardUsingCardNames  = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_name IN (%s)"
	queryCardsUsingProductID = "SELECT DISTINCT(card_number),card_color,card_name,card_attribute,card_effect,monster_type,monster_attack,monster_defense FROM product_contents WHERE product_id= ? ORDER BY card_name"
	queryRandomCardID        = "SELECT card_number FROM card_info WHERE card_color != 'Token' ORDER BY RAND() LIMIT 1"

	findRelatedCardsUsingCardEffect string = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE (card_effect LIKE ? OR card_effect LIKE ?) AND card_number != ? ORDER BY color_id, card_name"
)

// interface
type SKCDatabaseAccessObject interface {
	GetSKCDBVersion() (string, error)

	GetCardColorIDs() (map[string]int, *model.APIError)

	GetDesiredCardInDBUsingID(cardID string) (model.Card, *model.APIError)
	GetDesiredCardInDBUsingMultipleCardIDs(cards []string) (*model.BatchCardData[model.CardIDs], *model.APIError)
	GetDesiredCardsFromDBUsingMultipleCardNames(cardName []string) (*model.BatchCardData[model.CardNames], *model.APIError)
	GetCardsFoundInProduct(productID string) (*model.BatchCardData[model.CardIDs], *model.APIError)

	GetOccurrenceOfCardNameInAllCardEffect(cardName string, cardId string) ([]model.Card, *model.APIError)

	GetInArchetypeSupportUsingCardName(archetypeName string) ([]model.Card, *model.APIError)
	GetInArchetypeSupportUsingCardText(archetypeName string) ([]model.Card, *model.APIError)
	GetArchetypeExclusionsUsingCardText(archetypeName string) ([]model.Card, *model.APIError)

	GetDesiredProductInDBUsingMultipleProductIDs(cards []string) (*model.BatchProductData[model.ProductIDs], *model.APIError)

	GetRandomCard() (string, *model.APIError)
}

// impl
type SKCDAOImplementation struct{}

// Get version of MYSQL being used by SKC DB.
func (imp SKCDAOImplementation) GetSKCDBVersion() (string, error) {
	var version string
	if err := skcDBConn.QueryRow(queryDBVersion).Scan(&version); err != nil {
		log.Printf("Error getting SKC DB version - %v", err)
		return version, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	}

	return version, nil
}

// Get IDs for all card colors currently in database.
func (imp SKCDAOImplementation) GetCardColorIDs() (map[string]int, *model.APIError) {
	log.Println("Retrieving card color IDs from DB")
	cardColorIDs := map[string]int{}

	if rows, err := skcDBConn.Query(queryCardColorIDs); err != nil {
		log.Printf(queryErrorLog, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		for rows.Next() {
			var colorId int
			var cardColor string

			if err := rows.Scan(&colorId, &cardColor); err != nil {
				log.Printf(parseErrorLog, err)
				return cardColorIDs, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
			}

			cardColorIDs[cardColor] = colorId
		}
	}
	return cardColorIDs, nil
}

// Leverages GetDesiredCardInDBUsingMultipleCardIDs to get information on a specific card using its identifier
func (imp SKCDAOImplementation) GetDesiredCardInDBUsingID(cardID string) (model.Card, *model.APIError) {
	if results, err := imp.GetDesiredCardInDBUsingMultipleCardIDs([]string{cardID}); err != nil {
		return model.Card{}, err
	} else {
		if card, exists := results.CardInfo[cardID]; !exists {
			return model.Card{}, &model.APIError{Message: fmt.Sprintf("No results found when querying by card ID %s", cardID), StatusCode: http.StatusNotFound}
		} else {
			return card, nil
		}
	}
}

func (imp SKCDAOImplementation) GetDesiredCardInDBUsingMultipleCardIDs(cardIDs []string) (*model.BatchCardData[model.CardIDs], *model.APIError) {
	log.Printf("Retrieving card data from DB for cards w/ IDs %v", cardIDs)

	numCards := len(cardIDs)
	args := make([]interface{}, numCards)
	cardData := make(model.CardDataMap, numCards) // used to store results

	for index, cardId := range cardIDs {
		args[index] = cardId
	}

	query := fmt.Sprintf(queryCardUsingCardIDs, variablePlaceholders(numCards))

	if rows, err := skcDBConn.Query(query, args...); err != nil {
		log.Printf(queryErrorLog, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		if cards, err := parseRowsForCard(rows); err != nil {
			return nil, err
		} else {
			for _, card := range cards {
				cardData[card.CardID] = card
			}
		}
	}

	return &model.BatchCardData[model.CardIDs]{CardInfo: cardData, UnknownResources: cardData.FindMissingIDs(cardIDs)}, nil
}

func (imp SKCDAOImplementation) GetDesiredProductInDBUsingMultipleProductIDs(products []string) (*model.BatchProductData[model.ProductIDs], *model.APIError) {
	log.Printf("Retrieving product data from DB for product w/ IDs %v", products)

	numProducts := len(products)
	args := make([]interface{}, numProducts)
	productData := make(model.ProductDataMap, numProducts)

	for index, cardId := range products {
		args[index] = cardId
	}

	query := fmt.Sprintf("SELECT product_id, product_locale, product_name, product_release_date, product_content_total, product_type, product_sub_type FROM product_info WHERE product_id IN (%s)", variablePlaceholders(numProducts))

	if rows, err := skcDBConn.Query(query, args...); err != nil {
		log.Printf(queryErrorLog, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		for rows.Next() {
			var product model.Product
			if err := rows.Scan(&product.ProductID, &product.ProductLocale,
				&product.ProductName, &product.ProductReleaseDate, &product.ProductTotal, &product.ProductType, &product.ProductSubType); err != nil {
				log.Printf(parseErrorLog, err)
				return nil, &model.APIError{Message: "Error parsing data from DB.", StatusCode: http.StatusInternalServerError}
			}

			productData[product.ProductID] = product
		}
	}

	return &model.BatchProductData[model.ProductIDs]{ProductInfo: productData, UnknownResources: productData.FindMissingIDs(products)}, nil
}

// Uses card names to find instance of card
func (imp SKCDAOImplementation) GetDesiredCardsFromDBUsingMultipleCardNames(cardNames []string) (*model.BatchCardData[model.CardNames], *model.APIError) {
	log.Printf("Retrieving card data from DB for cards w/ name %v", cardNames)

	numCards := len(cardNames)
	args := make([]interface{}, numCards)
	cardData := make(model.CardDataMap, numCards) // used to store results

	for index, cardId := range cardNames {
		args[index] = cardId
	}

	query := fmt.Sprintf(queryCardUsingCardNames, variablePlaceholders(numCards))

	if rows, err := skcDBConn.Query(query, args...); err != nil {
		log.Printf(queryErrorLog, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		if cards, err := parseRowsForCard(rows); err != nil {
			return nil, err
		} else {
			for _, card := range cards {
				cardData[card.CardName] = card
			}
		}
	}

	return &model.BatchCardData[model.CardNames]{CardInfo: cardData, UnknownResources: cardData.FindMissingNames(cardNames)}, nil
}

// Uses card names to find instance of card
func (imp SKCDAOImplementation) GetCardsFoundInProduct(productId string) (*model.BatchCardData[model.CardIDs], *model.APIError) {
	log.Printf("Retrieving card data from DB found in product %v", productId)

	cardData := make(model.CardDataMap) // used to store results

	if rows, err := skcDBConn.Query(queryCardsUsingProductID, productId); err != nil {
		log.Printf(queryErrorLog, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		if cards, err := parseRowsForCard(rows); err != nil {
			return nil, err
		} else {
			for _, card := range cards {
				cardData[card.CardName] = card
			}
		}
	}

	return &model.BatchCardData[model.CardIDs]{CardInfo: cardData}, nil
}

// TODO: document
// TODO: find way to make code more readable
func (imp SKCDAOImplementation) GetOccurrenceOfCardNameInAllCardEffect(cardName string, cardId string) ([]model.Card, *model.APIError) {
	log.Printf("Retrieving card data from DB for all cards that reference card %s in their text", cardName)

	cardNameWithDoubleQuotes := `%"` + cardName + `"%`
	cardNameWithSingleQuotes := `%'` + cardName + `'%`

	if rows, err := skcDBConn.Query(findRelatedCardsUsingCardEffect, cardNameWithDoubleQuotes, cardNameWithSingleQuotes, cardId); err != nil {
		log.Printf(queryErrorLog, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		return parseRowsForCard(rows)
	}
}

func (imp SKCDAOImplementation) GetInArchetypeSupportUsingCardName(archetypeName string) ([]model.Card, *model.APIError) {
	log.Printf("Retrieving card data from DB for all cards that reference archetype %s in their name", archetypeName)
	searchTerm := `%` + archetypeName + `%`

	if rows, err := skcDBConn.Query("SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_name LIKE BINARY ? ORDER BY card_name", searchTerm); err != nil {
		log.Printf(queryErrorLog, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		return parseRowsForCard(rows)
	}
}

func (imp SKCDAOImplementation) GetInArchetypeSupportUsingCardText(archetypeName string) ([]model.Card, *model.APIError) {
	log.Printf("Retrieving card data from DB for all cards treated as archetype %s", archetypeName)
	archetypeName = `%` + fmt.Sprintf(`This card is always treated as an "%s" card`, archetypeName) + `%`

	if rows, err := skcDBConn.Query("SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_effect LIKE ? ORDER BY card_name", archetypeName); err != nil {
		log.Printf(queryErrorLog, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		return parseRowsForCard(rows)
	}
}

func (imp SKCDAOImplementation) GetArchetypeExclusionsUsingCardText(archetypeName string) ([]model.Card, *model.APIError) {
	log.Printf("Retrieving card data from DB for all cards explicitly not treated as archetype %s", archetypeName)
	archetypeName = `%` + fmt.Sprintf(`This card is not treated as a "%s" card`, archetypeName) + `%`

	if rows, err := skcDBConn.Query("SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_effect LIKE ? ORDER BY card_name", archetypeName); err != nil {
		log.Printf(queryErrorLog, err)
		return nil, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	} else {
		return parseRowsForCard(rows)
	}
}

func (imp SKCDAOImplementation) GetRandomCard() (string, *model.APIError) {
	var randomCardId string

	if err := skcDBConn.QueryRow(queryRandomCardID).Scan(&randomCardId); err != nil {
		log.Printf(queryErrorLog, err)
		return "", &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	}
	return randomCardId, nil
}

func parseRowsForCard(rows *sql.Rows) ([]model.Card, *model.APIError) {
	cards := []model.Card{}

	for rows.Next() {
		var card model.Card
		if err := rows.Scan(&card.CardID, &card.CardColor, &card.CardName, &card.CardAttribute, &card.CardEffect, &card.MonsterType, &card.MonsterAttack, &card.MonsterDefense); err != nil {
			log.Printf(parseErrorLog, err)
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
