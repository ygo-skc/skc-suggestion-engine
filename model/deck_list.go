package model

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ygo-skc/skc-suggestion-engine/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeckList struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	Name              string             `bson:"name" json:"name" validate:"required,decklistname"`
	ContentB64        string             `bson:"content" json:"listContent" validate:"required,base64"`
	VideoUrl          string             `bson:"videoUrl" validate:"omitempty,url"`
	DeckMascots       []string           `bson:"deckMascots" json:"deckMascots" validate:"omitempty,deckmascots"`
	NumMainDeckCards  int                `bson:"numMainDeckCards" json:"numMainDeckCards"`
	NumExtraDeckCards int                `bson:"numExtraDeckCards" json:"numExtraDeckCards"`
	Tags              []string           `bson:"tags" json:"tags" validate:"required"`
	CreatedAt         time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt         time.Time          `bson:"updatedAt" json:"updatedAt"`
	Content           *[]Content         `bson:"omitempty" json:"content"`
}

type Content struct {
	Quantity int  `bson:"omitempty" json:"quantity"`
	Card     Card `bson:"omitempty" json:"card"`
}

type DeckListContents map[string]Card

type DeckListBreakdown struct {
	CardQuantity           map[string]int
	CardIDs                []string
	InvalidIDs             []string
	AllCards               DeckListContents
	MainDeck               DeckListContents
	ExtraDeck              DeckListContents
	MainDeckCardIDsSorted  []string
	ExtraDeckCardIDsSorted []string
	NumMainDeckCards       int
	NumExtraDeckCards      int
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

	dlb.NumMainDeckCards = numMainDeckCards
	dlb.NumExtraDeckCards = numExtraDeckCards

	dlb.MainDeckCardIDsSorted = sortDeckUsingName(dlb.MainDeck)
	dlb.ExtraDeckCardIDsSorted = sortDeckUsingName(dlb.ExtraDeck)
}

func sortDeckUsingName(dlc DeckListContents) []string {
	sortedIDs := make([]string, 0, len(dlc))

	for id := range dlc {
		sortedIDs = append(sortedIDs, id)
	}
	sort.SliceStable(sortedIDs, func(i, j int) bool {
		return dlc[sortedIDs[i]].CardName < dlc[sortedIDs[j]].CardName
	})

	return sortedIDs
}

func (dlb DeckListBreakdown) ListStringCleanup() string {
	formattedDLS := "Main Deck\n"

	for _, cardID := range dlb.MainDeckCardIDsSorted {
		formattedDLS += formattedLine(dlb.MainDeck, cardID, dlb.CardQuantity[cardID])
	}

	formattedDLS += "\nExtra Deck\n"

	for _, cardID := range dlb.ExtraDeckCardIDsSorted {
		formattedDLS += formattedLine(dlb.ExtraDeck, cardID, dlb.CardQuantity[cardID])
	}

	return formattedDLS
}

func formattedLine(dlc DeckListContents, cardID string, quantity int) string {
	return fmt.Sprintf("%dx%s|%s\n", quantity, cardID, dlc[cardID].CardName)
}

func (dlb DeckListBreakdown) Validate() APIError {
	if len(dlb.InvalidIDs) > 0 {
		log.Println("Deck list contains card(s) that were not found in skc DB. All cards not found in DB:", dlb.InvalidIDs)
		return APIError{Message: "Found cards in deck list that are not yet in the database. Remove the cards before submitting again. Cards not found " + strings.Join(dlb.InvalidIDs, ", ")}
	}

	// validate extra deck has correct number of cards
	if dlb.NumExtraDeckCards > 15 {
		log.Println("Extra deck cannot contain more than 15 cards. Found", dlb.NumExtraDeckCards)
		return APIError{Message: "Too many extra deck cards found in deck list. Found " + strconv.Itoa(dlb.NumExtraDeckCards)}
	}

	// validate main deck has correct number of cards
	if dlb.NumMainDeckCards < 40 || dlb.NumMainDeckCards > 60 {
		log.Printf("Main deck cannot contain less than 40 cards and no more than 60 cards. Found %d.", dlb.NumMainDeckCards)
		return APIError{Message: "Main deck cannot contain less than 40 cards and cannot contain more than 60 cards. Found " + strconv.Itoa(dlb.NumMainDeckCards) + "."}
	}

	return APIError{}
}
