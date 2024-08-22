package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
)

var (
	skcDBConn  *sql.DB
	spaceRegex = regexp.MustCompile(`[ ]+`)
)

const (
	// errors
	genericError = "Error occurred while querying DB"

	// queries
	queryDBVersion    = "SELECT VERSION()"
	queryCardColorIDs = "SELECT color_id, card_color from card_colors ORDER BY color_id"

	queryCardUsingCardID     = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_number = ?"
	queryCardUsingCardIDs    = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_number IN (%s)"
	queryCardUsingCardNames  = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_name IN (%s)"
	queryCardsUsingProductID = "SELECT DISTINCT(card_number), card_color,card_name,card_attribute,card_effect,monster_type,monster_attack,monster_defense FROM product_contents WHERE product_id= ? ORDER BY card_name"
	queryRandomCardID        = "SELECT card_number FROM card_info WHERE card_color != 'Token' ORDER BY RAND() LIMIT 1"

	queryCardsInArchetypeUsingName  = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE card_name LIKE BINARY ? ORDER BY card_name"
	queryCardsTreatedAsArchetype    = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE MATCH(card_effect) AGAINST(? IN BOOLEAN MODE) ORDER BY card_name"
	queryCardsNotTreatedAsArchetype = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE MATCH(card_effect) AGAINST(? IN BOOLEAN MODE) ORDER BY card_name"

	findRelatedCardsUsingCardEffect = "SELECT card_number, card_color, card_name, card_attribute, card_effect, monster_type, monster_attack, monster_defense FROM card_info WHERE MATCH(card_effect) AGAINST(? IN BOOLEAN MODE) AND card_number != ? ORDER BY color_id, card_name"
)

func convertToFullText(subject string) string {
	fullTextSubject := spaceRegex.ReplaceAllString(strings.ReplaceAll(subject, "-", " "), " +")
	return fmt.Sprintf(`"+%s"`, fullTextSubject) // match phrase, not all words in text will match only consecutive matches of words in phrase
}

func buildVariableQuerySubjects(subjects []string) ([]interface{}, int) {
	numSubjects := len(subjects)
	args := make([]interface{}, numSubjects)

	for index, cardId := range subjects {
		args[index] = cardId
	}

	return args, numSubjects
}

func variablePlaceholders(totalFields int) string {
	switch totalFields {
	case 0:
		return ""
	case 1:
		return "?"
	default:
		return fmt.Sprintf("?%s", strings.Repeat(", ?", totalFields-1))
	}
}

func handleQueryError(logger *slog.Logger, err error) *model.APIError {
	logger.Error(fmt.Sprintf("Error fetching data from DB - %v", err))
	return &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
}

func handleRowParsingError(logger *slog.Logger, err error) *model.APIError {
	logger.Error(fmt.Sprintf("Error parsing data from DB - %v", err))
	return &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
}

// interface
type SKCDatabaseAccessObject interface {
	GetSKCDBVersion(context.Context) (string, error)

	GetCardColorIDs(context.Context) (map[string]int, *model.APIError)

	GetDesiredCardInDBUsingID(context.Context, string) (model.Card, *model.APIError)
	GetDesiredCardInDBUsingMultipleCardIDs(context.Context, []string) (model.BatchCardData[model.CardIDs], *model.APIError)
	GetDesiredCardsFromDBUsingMultipleCardNames(context.Context, []string) (model.BatchCardData[model.CardNames], *model.APIError)
	GetCardsFoundInProduct(context.Context, string) (model.BatchCardData[model.CardIDs], *model.APIError)

	GetOccurrenceOfCardNameInAllCardEffect(context.Context, string, string) ([]model.Card, *model.APIError)

	GetInArchetypeSupportUsingCardName(context.Context, string) ([]model.Card, *model.APIError)
	GetInArchetypeSupportUsingCardText(context.Context, string) ([]model.Card, *model.APIError)
	GetArchetypeExclusionsUsingCardText(context.Context, string) ([]model.Card, *model.APIError)

	GetDesiredProductInDBUsingMultipleProductIDs(context.Context, []string) (model.BatchProductData[model.ProductIDs], *model.APIError)

	GetRandomCard(context.Context) (string, *model.APIError)
}

// impl
type SKCDAOImplementation struct{}

// Get version of MYSQL being used by SKC DB.
func (imp SKCDAOImplementation) GetSKCDBVersion(ctx context.Context) (string, error) {
	var version string
	if err := skcDBConn.QueryRow(queryDBVersion).Scan(&version); err != nil {
		util.LoggerFromContext(ctx).Info(fmt.Sprintf("Error getting SKC DB version - %v", err))
		return version, &model.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	}

	return version, nil
}

// Get IDs for all card colors currently in database.
func (imp SKCDAOImplementation) GetCardColorIDs(ctx context.Context) (map[string]int, *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	logger.Info("Retrieving card color IDs from DB")
	cardColorIDs := map[string]int{}

	if rows, err := skcDBConn.Query(queryCardColorIDs); err != nil {
		return nil, handleQueryError(logger, err)
	} else {
		for rows.Next() {
			var colorId int
			var cardColor string

			if err := rows.Scan(&colorId, &cardColor); err != nil {
				return cardColorIDs, handleRowParsingError(logger, err)
			}

			cardColorIDs[cardColor] = colorId
		}
	}
	return cardColorIDs, nil
}

// Leverages GetDesiredCardInDBUsingMultipleCardIDs to get information on a specific card using its identifier
func (imp SKCDAOImplementation) GetDesiredCardInDBUsingID(ctx context.Context, cardID string) (model.Card, *model.APIError) {
	if results, err := imp.GetDesiredCardInDBUsingMultipleCardIDs(ctx, []string{cardID}); err != nil {
		return model.Card{}, err
	} else {
		if card, exists := results.CardInfo[cardID]; !exists {
			return model.Card{}, &model.APIError{Message: fmt.Sprintf("No results found when querying by card ID %s", cardID), StatusCode: http.StatusNotFound}
		} else {
			return card, nil
		}
	}
}

func (imp SKCDAOImplementation) GetDesiredCardInDBUsingMultipleCardIDs(ctx context.Context, cardIDs []string) (model.BatchCardData[model.CardIDs], *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	logger.Info("Retrieving card data from DB")

	args, numCards := buildVariableQuerySubjects(cardIDs)
	cardData := make(model.CardDataMap, numCards) // used to store results

	query := fmt.Sprintf(queryCardUsingCardIDs, variablePlaceholders(numCards))

	if rows, err := skcDBConn.Query(query, args...); err != nil {
		return model.BatchCardData[model.CardIDs]{}, handleQueryError(logger, err)
	} else {
		if cards, err := parseRowsForCard(ctx, rows); err != nil {
			return model.BatchCardData[model.CardIDs]{}, err
		} else {
			for _, card := range cards {
				cardData[card.CardID] = card
			}
		}
	}

	return model.BatchCardData[model.CardIDs]{CardInfo: cardData, UnknownResources: cardData.FindMissingIDs(cardIDs)}, nil
}

func (imp SKCDAOImplementation) GetDesiredProductInDBUsingMultipleProductIDs(ctx context.Context, products []string) (model.BatchProductData[model.ProductIDs], *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	logger.Info("Retrieving product data from DB")

	args, numProducts := buildVariableQuerySubjects(products)
	productData := make(model.ProductDataMap, numProducts)

	query := fmt.Sprintf("SELECT product_id, product_locale, product_name, product_release_date, product_content_total, product_type, product_sub_type FROM product_info WHERE product_id IN (%s)", variablePlaceholders(numProducts))

	if rows, err := skcDBConn.Query(query, args...); err != nil {
		return model.BatchProductData[model.ProductIDs]{}, handleQueryError(logger, err)
	} else {
		for rows.Next() {
			var product model.Product
			if err := rows.Scan(&product.ProductID, &product.ProductLocale,
				&product.ProductName, &product.ProductReleaseDate, &product.ProductTotal, &product.ProductType, &product.ProductSubType); err != nil {
				return model.BatchProductData[model.ProductIDs]{}, handleRowParsingError(logger, err)
			}

			productData[product.ProductID] = product
		}
	}

	return model.BatchProductData[model.ProductIDs]{ProductInfo: productData, UnknownResources: productData.FindMissingIDs(products)}, nil
}

// Uses card names to find instance of card
func (imp SKCDAOImplementation) GetDesiredCardsFromDBUsingMultipleCardNames(ctx context.Context, cardNames []string) (model.BatchCardData[model.CardNames], *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	logger.Info(fmt.Sprintf("Retrieving card data from DB for cards w/ name %v", cardNames))

	args, numCards := buildVariableQuerySubjects(cardNames)
	cardData := make(model.CardDataMap, numCards) // used to store results

	query := fmt.Sprintf(queryCardUsingCardNames, variablePlaceholders(numCards))

	if rows, err := skcDBConn.Query(query, args...); err != nil {
		return model.BatchCardData[model.CardNames]{}, handleQueryError(logger, err)
	} else {
		if cards, err := parseRowsForCard(ctx, rows); err != nil {
			return model.BatchCardData[model.CardNames]{}, err
		} else {
			for _, card := range cards {
				cardData[card.CardName] = card
			}
		}
	}

	return model.BatchCardData[model.CardNames]{CardInfo: cardData, UnknownResources: cardData.FindMissingNames(cardNames)}, nil
}

// Uses card names to find instance of card
func (imp SKCDAOImplementation) GetCardsFoundInProduct(ctx context.Context, productId string) (model.BatchCardData[model.CardIDs], *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	logger.Info("Retrieving cards found in product")

	cardData := make(model.CardDataMap) // used to store results

	if rows, err := skcDBConn.Query(queryCardsUsingProductID, productId); err != nil {
		return model.BatchCardData[model.CardIDs]{}, handleQueryError(logger, err)
	} else {
		if cards, err := parseRowsForCard(ctx, rows); err != nil {
			return model.BatchCardData[model.CardIDs]{}, err
		} else {
			for _, card := range cards {
				cardData[card.CardID] = card
			}
		}
	}

	return model.BatchCardData[model.CardIDs]{CardInfo: cardData}, nil
}

// TODO: document
// TODO: find way to make code more readable
func (imp SKCDAOImplementation) GetOccurrenceOfCardNameInAllCardEffect(ctx context.Context, cardName string, cardId string) ([]model.Card, *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	logger.Info(fmt.Sprintf("Retrieving card data from DB for all cards that reference card %s in their text", cardName))

	if rows, err := skcDBConn.Query(findRelatedCardsUsingCardEffect, convertToFullText(cardName), cardId); err != nil {
		return nil, handleQueryError(logger, err)
	} else {
		return parseRowsForCard(ctx, rows)
	}
}

func (imp SKCDAOImplementation) GetInArchetypeSupportUsingCardName(ctx context.Context, archetypeName string) ([]model.Card, *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	logger.Info("Retrieving card data from DB for all cards that reference archetype in their name")
	searchTerm := `%` + archetypeName + `%`

	if rows, err := skcDBConn.Query(queryCardsInArchetypeUsingName, searchTerm); err != nil {
		return nil, handleQueryError(logger, err)
	} else {
		return parseRowsForCard(ctx, rows)
	}
}

func (imp SKCDAOImplementation) GetInArchetypeSupportUsingCardText(ctx context.Context, archetypeName string) ([]model.Card, *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	logger.Info("Retrieving card data from DB for all cards treated as archetype")

	if rows, err := skcDBConn.Query(queryCardsTreatedAsArchetype, fmt.Sprintf(`+"This card is always treated as" +"%s card"`, archetypeName)); err != nil {
		return nil, handleQueryError(logger, err)
	} else {
		return parseRowsForCard(ctx, rows)
	}
}

func (imp SKCDAOImplementation) GetArchetypeExclusionsUsingCardText(ctx context.Context, archetypeName string) ([]model.Card, *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	logger.Info("Retrieving card data from DB for all cards explicitly not treated as archetype")

	if rows, err := skcDBConn.Query(queryCardsNotTreatedAsArchetype, fmt.Sprintf(`+"This card is not treated as" +"%s card"`, archetypeName)); err != nil {
		return nil, handleQueryError(logger, err)
	} else {
		return parseRowsForCard(ctx, rows)
	}
}

func (imp SKCDAOImplementation) GetRandomCard(ctx context.Context) (string, *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	var randomCardId string

	if err := skcDBConn.QueryRow(queryRandomCardID).Scan(&randomCardId); err != nil {
		return "", handleQueryError(logger, err)
	}
	return randomCardId, nil
}

func parseRowsForCard(ctx context.Context, rows *sql.Rows) ([]model.Card, *model.APIError) {
	logger := util.LoggerFromContext(ctx)
	cards := []model.Card{}

	for rows.Next() {
		var card model.Card
		if err := rows.Scan(&card.CardID, &card.CardColor, &card.CardName, &card.CardAttribute, &card.CardEffect,
			&card.MonsterType, &card.MonsterAttack, &card.MonsterDefense); err != nil {
			return nil, handleRowParsingError(logger, err)
		} else {
			cards = append(cards, card)
		}
	}

	return cards, nil // no parsing error
}
