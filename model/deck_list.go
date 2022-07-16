package model

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ygo-skc/skc-suggestion-engine/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeckList struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name" json:"name" validate:"required,decklistname"`
	ListContent string             `bson:"contents" json:"listContent" validate:"required,base64"`
	VideoUrl    string             `bson:"videoUrl" validate:"omitempty,url"`
	Tags        []string           `bson:"tags" json:"tags" validate:"required"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type DeckListContents map[string]Card

type DeckListBreakdown struct {
	CardQuantity map[string]int
	CardIDs      []string
}

// validate and handle validation error messages
func (dl DeckList) Validate() util.APIError {
	if err := util.V.Struct(dl); err != nil {
		errMessages := []string{}
		for _, e := range err.(validator.ValidationErrors) {
			errMessages = append(errMessages, e.Translate(util.Translator))
		}

		message := strings.Join(errMessages, " ")
		log.Println("There were", len(errMessages), "errors while validating input. Errors:", message)

		return util.APIError{Message: message}
	}

	return util.APIError{}
}

func (dls DeckListContents) Validate(cardCopiesInDeck map[string]int, idsForCardsInDeckList []string) util.APIError {
	invalidIDs := []string{}
	mainDeckCards := []string{}
	extraDeckCards := []string{}
	for _, cardID := range idsForCardsInDeckList {
		if _, isPresent := dls[cardID]; !isPresent {
			invalidIDs = append(invalidIDs, cardID)
		} else {
			if dls[cardID].isExtraDeckMonster() {
				extraDeckCards = append(extraDeckCards, dls[cardID].CardID)
			} else {
				mainDeckCards = append(mainDeckCards, dls[cardID].CardID)
			}
		}
	}

	if len(invalidIDs) > 0 {
		log.Println("Deck list contains card(s) that were not found in skc DB. All cards not found in DB:", invalidIDs)
		return util.APIError{Message: "Found cards in deck list that are not yet in the database. Remove the cards before submitting again. Cards not found " + strings.Join(invalidIDs, ", ")}
	}

	numExtraDeckCards := len(extraDeckCards)
	if numExtraDeckCards > 15 {
		log.Println("Extra deck cannot contain more than 15 cards. Found", numExtraDeckCards)
		return util.APIError{Message: "Too many extra deck cards found in deck list. Found " + strconv.Itoa(numExtraDeckCards)}
	}

	numMainDeckCards := len(mainDeckCards)
	if numMainDeckCards < 40 || numMainDeckCards > 60 {
		log.Printf("Main deck cannot contain less than 40 cards and no more than 60 cards. Found %d.", numMainDeckCards)
		return util.APIError{Message: "Main deck cannot contain less than 40 cards and cannot contain more than 60 cards. Found " + strconv.Itoa(numMainDeckCards) + "."}
	}

	return util.APIError{}
}
