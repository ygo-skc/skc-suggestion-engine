package model

type CardDataMap map[string]Card

// finds all card IDs not found in CardDataMap keys
func (cardData CardDataMap) FindMissingIDs(cardIDs CardIDs) CardIDs {
	missingIDs := make(CardIDs, 0)

	for _, cardID := range cardIDs {
		if _, containsKey := cardData[cardID]; !containsKey {
			missingIDs = append(missingIDs, cardID)
		}
	}

	return missingIDs
}

type ProductDataMap map[string]Product

// finds all product IDs not found in ProductDataMap keys
func (productData ProductDataMap) FindMissingIDs(productIDs ProductIDs) ProductIDs {
	missingIDs := make(ProductIDs, 0)

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
