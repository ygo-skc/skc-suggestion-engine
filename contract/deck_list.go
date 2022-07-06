package contract

import (
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeckList struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name" validate:"required,decklistname"`
	ListContent string             `bson:"contents" validate:"required,base64"`
	VideoUrl    string             `bson:"videoUrl" validate:"omitempty,url"`
	Tags        []string           `bson:"tags" validate:"required"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type DeckListContents map[string]Card
type DeckListBreakdown struct {
	CardQuantity map[string]int
	CardIDs      []string
}

func (dls DeckListContents) Validate(cardCopiesInDeck map[string]int, idsForCardsInDeckList []string) error {
	invalidIDs := []string{}
	for _, cardID := range idsForCardsInDeckList {
		if _, isPresent := dls[cardID]; !isPresent {
			invalidIDs = append(invalidIDs, cardID)
		}
	}

	if len(invalidIDs) > 0 {
		log.Println("Deck list contains card(s) that were not found in skc DB. All cards not found in DB:", invalidIDs)
		return errors.New("422")
	}

	return nil
}
