package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
)

var (
	skcDBConn  *sql.DB
	spaceRegex = regexp.MustCompile(`[ ]+`)
)

const (
	// errors
	genericError = "Error occurred while querying DB"

	// queries
	queryDBVersion = "SELECT VERSION()"

	queryCardsUsingProductID = "SELECT DISTINCT(card_number), card_color,card_name,card_attribute,card_effect,monster_type,monster_attack,monster_defense FROM product_contents WHERE product_id= ? ORDER BY card_name"

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

func handleQueryError(logger *slog.Logger, err error) *cModel.APIError {
	logger.Error(fmt.Sprintf("Error fetching data from DB - %v", err))
	return &cModel.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
}

func handleRowParsingError(logger *slog.Logger, err error) *cModel.APIError {
	logger.Error(fmt.Sprintf("Error parsing data from DB - %v", err))
	return &cModel.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
}

// interface
type SKCDatabaseAccessObject interface {
	GetSKCDBVersion(context.Context) (string, error)

	GetCardsFoundInProduct(context.Context, string) (cModel.BatchCardData[cModel.CardIDs], *cModel.APIError)

	GetOccurrenceOfCardNameInAllCardEffect(context.Context, string, string) ([]cModel.YGOCard, *cModel.APIError)

	GetInArchetypeSupportUsingCardText(context.Context, string) ([]cModel.YGOCard, *cModel.APIError)
	GetArchetypeExclusionsUsingCardText(context.Context, string) ([]cModel.YGOCard, *cModel.APIError)

	GetDesiredProductInDBUsingID(context.Context, string) (*cModel.YGOProduct, *cModel.APIError)
	GetDesiredProductInDBUsingMultipleProductIDs(context.Context, []string) (cModel.BatchProductData[cModel.ProductIDs], *cModel.APIError)
}

// impl
type SKCDAOImplementation struct{}

// Get version of MYSQL being used by SKC DB.
func (imp SKCDAOImplementation) GetSKCDBVersion(ctx context.Context) (string, error) {
	var version string
	if err := skcDBConn.QueryRow(queryDBVersion).Scan(&version); err != nil {
		cUtil.LoggerFromContext(ctx).Info(fmt.Sprintf("Error getting SKC DB version - %v", err))
		return version, &cModel.APIError{Message: genericError, StatusCode: http.StatusInternalServerError}
	}

	return version, nil
}

// Leverages GetDesiredProductInDBUsingMultipleProductIDs to get information on a specific product using its identifier
func (imp SKCDAOImplementation) GetDesiredProductInDBUsingID(ctx context.Context, productID string) (*cModel.YGOProduct, *cModel.APIError) {
	if results, err := imp.GetDesiredProductInDBUsingMultipleProductIDs(ctx, []string{productID}); err != nil {
		return nil, err
	} else {
		if product, exists := results.ProductInfo[productID]; !exists {
			return nil, &cModel.APIError{Message: fmt.Sprintf("No results found when querying by product ID %s", productID), StatusCode: http.StatusNotFound}
		} else {
			return &product, nil
		}
	}
}

func (imp SKCDAOImplementation) GetDesiredProductInDBUsingMultipleProductIDs(ctx context.Context, products []string) (cModel.BatchProductData[cModel.ProductIDs], *cModel.APIError) {
	logger := cUtil.LoggerFromContext(ctx)
	logger.Info("Retrieving product data from DB")

	args, numProducts := buildVariableQuerySubjects(products)
	productData := make(cModel.ProductDataMap, numProducts)

	query := fmt.Sprintf("SELECT product_id, product_locale, product_name, product_release_date, product_content_total, product_type, product_sub_type FROM product_info WHERE product_id IN (%s)", variablePlaceholders(numProducts))

	if rows, err := skcDBConn.Query(query, args...); err != nil {
		return cModel.BatchProductData[cModel.ProductIDs]{}, handleQueryError(logger, err)
	} else {
		for rows.Next() {
			var product cModel.YGOProductREST
			if err := rows.Scan(&product.ID, &product.Locale, &product.Name, &product.ReleaseDate, &product.Total, &product.Type, &product.SubType); err != nil {
				return cModel.BatchProductData[cModel.ProductIDs]{}, handleRowParsingError(logger, err)
			}

			productData[product.ID] = product
		}
	}

	return cModel.BatchProductData[cModel.ProductIDs]{ProductInfo: productData, UnknownResources: productData.FindMissingIDs(products)}, nil
}

// Uses card names to find instance of card
func (imp SKCDAOImplementation) GetCardsFoundInProduct(ctx context.Context, productId string) (cModel.BatchCardData[cModel.CardIDs], *cModel.APIError) {
	logger := cUtil.LoggerFromContext(ctx)
	logger.Info("Retrieving cards found in product")

	cardData := make(cModel.CardDataMap) // used to store results

	if rows, err := skcDBConn.Query(queryCardsUsingProductID, productId); err != nil {
		return cModel.BatchCardData[cModel.CardIDs]{}, handleQueryError(logger, err)
	} else {
		if cards, err := parseRowsForCard(ctx, rows); err != nil {
			return cModel.BatchCardData[cModel.CardIDs]{}, err
		} else {
			for _, card := range cards {
				cardData[card.GetID()] = card
			}
		}
	}

	return cModel.BatchCardData[cModel.CardIDs]{CardInfo: cardData}, nil
}

// TODO: document
// TODO: find way to make code more readable
func (imp SKCDAOImplementation) GetOccurrenceOfCardNameInAllCardEffect(ctx context.Context, cardName string, cardId string) ([]cModel.YGOCard, *cModel.APIError) {
	logger := cUtil.LoggerFromContext(ctx)
	logger.Info(fmt.Sprintf("Retrieving card data from DB for all cards that reference card %s in their text", cardName))

	if rows, err := skcDBConn.Query(findRelatedCardsUsingCardEffect, convertToFullText(cardName), cardId); err != nil {
		return nil, handleQueryError(logger, err)
	} else {
		return parseRowsForCard(ctx, rows)
	}
}

func (imp SKCDAOImplementation) GetInArchetypeSupportUsingCardText(ctx context.Context, archetypeName string) ([]cModel.YGOCard, *cModel.APIError) {
	logger := cUtil.LoggerFromContext(ctx)
	logger.Info("Retrieving card data from DB for all cards treated as archetype")

	if rows, err := skcDBConn.Query(queryCardsTreatedAsArchetype, fmt.Sprintf(`+"This card is always treated as" +"%s card"`, archetypeName)); err != nil {
		return nil, handleQueryError(logger, err)
	} else {
		c, err := parseRowsForCard(ctx, rows)
		return c, err
	}
}

func (imp SKCDAOImplementation) GetArchetypeExclusionsUsingCardText(ctx context.Context, archetypeName string) ([]cModel.YGOCard, *cModel.APIError) {
	logger := cUtil.LoggerFromContext(ctx)
	logger.Info("Retrieving card data from DB for all cards explicitly not treated as archetype")

	if rows, err := skcDBConn.Query(queryCardsNotTreatedAsArchetype, fmt.Sprintf(`+"This card is not treated as" +"%s card"`, archetypeName)); err != nil {
		return nil, handleQueryError(logger, err)
	} else {
		c, err := parseRowsForCard(ctx, rows)
		return c, err
	}
}

func parseRowsForCard(ctx context.Context, rows *sql.Rows) ([]cModel.YGOCard, *cModel.APIError) {
	logger := cUtil.LoggerFromContext(ctx)
	cards := []cModel.YGOCard{}

	for rows.Next() {
		var card cModel.YGOCardREST
		if err := rows.Scan(&card.ID, &card.Color, &card.Name, &card.Attribute, &card.Effect, &card.MonsterType, &card.Attack, &card.Defense); err != nil {
			return nil, handleRowParsingError(logger, err)
		} else {
			cards = append(cards, card)
		}
	}

	return cards, nil // no parsing error
}
