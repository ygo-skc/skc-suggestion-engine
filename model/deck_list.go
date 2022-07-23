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
	DeckMascots []string           `bson:"deckMascots" json:"deckMascots" validate:"omitempty,deckmascots"`
	Tags        []string           `bson:"tags" json:"tags" validate:"required"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type DeckListContents map[string]Card

type DeckListBreakdown struct {
	CardQuantity      map[string]int
	CardIDs           []string
	InvalidIDs        []string
	AllCards          DeckListContents
	MainDeck          DeckListContents
	ExtraDeck         DeckListContents
	numMainDeckCards  int
	numExtraDeckCards int
}

// validate and handle validation error messages
func (dl DeckList) Validate() APIError {
	if err := util.V.Struct(dl); err != nil {
		errMessages := []string{}
		for _, e := range err.(validator.ValidationErrors) {
			errMessages = append(errMessages, e.Translate(util.Translator))
		}

		message := strings.Join(errMessages, " ")
		log.Println("There were", len(errMessages), "errors while validating input. Errors:", message)

		return APIError{Message: message}
	}

	return APIError{}
}

func (dlb *DeckListBreakdown) Sort() {
	dlb.MainDeck = map[string]Card{}
	dlb.ExtraDeck = map[string]Card{}
	numMainDeckCards := 0
	numExtraDeckCards := 0

	for _, cardID := range dlb.CardIDs {
		if _, isPresent := dlb.AllCards[cardID]; !isPresent {
			dlb.InvalidIDs = append(dlb.InvalidIDs, cardID)
		} else {
			if dlb.AllCards[cardID].isExtraDeckMonster() {
				dlb.ExtraDeck[cardID] = dlb.AllCards[cardID]
				numExtraDeckCards += dlb.CardQuantity[cardID]
			} else {
				dlb.MainDeck[cardID] = dlb.AllCards[cardID]
				numMainDeckCards += dlb.CardQuantity[cardID]
			}
		}
	}

	dlb.numMainDeckCards = numMainDeckCards
	dlb.numExtraDeckCards = numExtraDeckCards
}

func (dlb DeckListBreakdown) Validate() APIError {
	if len(dlb.InvalidIDs) > 0 {
		log.Println("Deck list contains card(s) that were not found in skc DB. All cards not found in DB:", dlb.InvalidIDs)
		return APIError{Message: "Found cards in deck list that are not yet in the database. Remove the cards before submitting again. Cards not found " + strings.Join(dlb.InvalidIDs, ", ")}
	}

	// validate extra deck has correct number of cards
	if dlb.numExtraDeckCards > 15 {
		log.Println("Extra deck cannot contain more than 15 cards. Found", dlb.numExtraDeckCards)
		return APIError{Message: "Too many extra deck cards found in deck list. Found " + strconv.Itoa(dlb.numExtraDeckCards)}
	}

	// validate main deck has correct number of cards
	if dlb.numMainDeckCards < 40 || dlb.numMainDeckCards > 60 {
		log.Printf("Main deck cannot contain less than 40 cards and no more than 60 cards. Found %d.", dlb.numMainDeckCards)
		return APIError{Message: "Main deck cannot contain less than 40 cards and cannot contain more than 60 cards. Found " + strconv.Itoa(dlb.numMainDeckCards) + "."}
	}

	return APIError{}
}
