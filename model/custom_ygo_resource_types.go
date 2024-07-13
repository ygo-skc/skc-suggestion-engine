package model

type QuotedToken = string

// identifier arrays

type CardIDs []string
type CardNames []string
type ProductIDs []string
type ProductNames []string
type IdentifierSlice interface {
	CardIDs | CardNames | ProductIDs | ProductNames
}

// data maps

type CardDataMap map[string]Card
type ProductDataMap map[string]Product

// finds all card IDs not found in CardDataMap keys
func (cardData CardDataMap) FindMissingIDs(cardIDs CardIDs) CardIDs {
	missingIDs := make(CardIDs, 0, len(cardData))

	for _, cardID := range cardIDs {
		if _, containsKey := cardData[cardID]; !containsKey {
			missingIDs = append(missingIDs, cardID)
		}
	}

	return missingIDs
}

// finds all card IDs not found in CardDataMap keys
func (cardData CardDataMap) FindMissingNames(cardNames CardNames) CardNames {
	missingNames := make(CardNames, 0, len(cardNames))

	for _, cardName := range cardNames {
		if _, containsKey := cardData[cardName]; !containsKey {
			missingNames = append(missingNames, cardName)
		}
	}

	return missingNames
}

// finds all product IDs not found in ProductDataMap keys
func (productData ProductDataMap) FindMissingIDs(productIDs ProductIDs) ProductIDs {
	missingIDs := make(ProductIDs, 0, len(productIDs))

	for _, productID := range productIDs {
		if _, containsKey := productData[productID]; !containsKey {
			missingIDs = append(missingIDs, productID)
		}
	}

	return missingIDs
}

type ResourceDataMap interface {
	CardDataMap | ProductDataMap
}

// data types that contain many resources of the same data type

type BatchCardIDs struct {
	CardIDs CardIDs `json:"cardIDs" validate:"required,ygocardids"`
}

type BatchProductIDs struct {
	ProductIDs ProductIDs `json:"productIDs"`
}

type BatchCardData[IS IdentifierSlice] struct {
	CardInfo         CardDataMap `json:"cardInfo"`
	UnknownResources IS          `json:"unknownResources"`
}

type BatchProductData[IS IdentifierSlice] struct {
	ProductInfo      ProductDataMap `json:"productInfo"`
	UnknownResources IS             `json:"unknownResources"`
}

type BatchData[IS IdentifierSlice] interface {
	BatchCardData[IS] | BatchProductData[IS]
}
