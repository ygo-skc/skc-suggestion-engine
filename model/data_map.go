package model

type CardDataMap map[string]Card

// finds all card IDs not found in CardDataMap keys
func (cardData CardDataMap) FindMissingIDs(cardIDs []string) []string {
	missingIDs := make([]string, 0)

	for _, cardID := range cardIDs {
		if _, containsKey := cardData[cardID]; !containsKey {
			missingIDs = append(missingIDs, cardID)
		}
	}

	return missingIDs
}

type ProductDataMap map[string]Product

type ResourceDataMap interface {
	CardDataMap | ProductDataMap
}
