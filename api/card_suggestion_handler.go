package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
)

var (
	quotedStringRegex  = regexp.MustCompile("^(\"[ \\w\\d-:@,'.]{3,}?\"|'[ \\w\\d-:@,'.]{3,}?')|[\\W](\"[ \\w\\d-:@,'.]{3,}?\"|'[ \\w\\d-:@,'.]{3,}?')")
	noBatchSuggestions = model.BatchCardSuggestions[model.CardIDs]{NamedMaterials: []model.CardReference{}, NamedReferences: []model.CardReference{}, MaterialArchetypes: []string{},
		ReferencedArchetypes: []string{}, UnknownResources: []string{}, FalsePositives: []string{}}
)

// Handler that will be used by suggestion endpoint.
// Will retrieve fusion, synchro, etc materials and other references if they are explicitly mentioned by name and their name exists in the DB.
func getCardSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]

	logger, ctx := util.NewRequestSetup(context.Background(), "card suggestions", slog.String("cardID", cardID))
	logger.Info("Card suggestions requested")

	if cardToGetSuggestionsFor, err := skcDBInterface.GetDesiredCardInDBUsingID(ctx, cardID); err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		ccIDs, _ := skcDBInterface.GetCardColorIDs(ctx) // retrieve card color IDs
		suggestions := getCardSuggestions(ctx, cardToGetSuggestionsFor, ccIDs)

		logger.Info(fmt.Sprintf("%s: %d unique material references - %d unique named references", cardToGetSuggestionsFor.CardName,
			len(suggestions.NamedMaterials), len(suggestions.NamedReferences)))

		json.NewEncoder(res).Encode(suggestions)
	}
}

func getCardSuggestions(ctx context.Context, cardToGetSuggestionsFor model.Card, ccIDs map[string]int) model.CardSuggestions {
	suggestions := model.CardSuggestions{Card: cardToGetSuggestionsFor}
	materialString := cardToGetSuggestionsFor.GetPotentialMaterialsAsString()

	wg := sync.WaitGroup{}

	// get materials if card is from extra deck
	if cardToGetSuggestionsFor.IsExtraDeckMonster() {
		wg.Add(2)
		go getMaterialRefs(ctx, &suggestions, materialString, ccIDs, &wg)
	} else {
		wg.Add(1)
		suggestions.NamedMaterials = []model.CardReference{}
		suggestions.MaterialArchetypes = []string{}

		util.LoggerFromContext(ctx).Debug("Not and extra deck monster")
	}
	go getNonMaterialRefs(ctx, &suggestions, cardToGetSuggestionsFor, materialString, ccIDs, &wg)

	wg.Wait()
	return suggestions
}

func getMaterialRefs(ctx context.Context, s *model.CardSuggestions, materialString string, ccIDs map[string]int, wg *sync.WaitGroup) {
	defer wg.Done()
	s.NamedMaterials, s.MaterialArchetypes = getReferences(ctx, materialString)
	sortCardReferences(&s.NamedMaterials, ccIDs)
}

// get named references - excludes materials
// will also check and remove self references
func getNonMaterialRefs(ctx context.Context, s *model.CardSuggestions, cardToGetSuggestionsFor model.Card, materialString string, ccIDs map[string]int, wg *sync.WaitGroup) {
	defer wg.Done()
	s.NamedReferences, s.ReferencedArchetypes = getReferences(ctx, strings.ReplaceAll(cardToGetSuggestionsFor.CardEffect, materialString, ""))
	s.HasSelfReference = util.RemoveSelfReference(cardToGetSuggestionsFor.CardName, &s.NamedReferences)
	sortCardReferences(&s.NamedReferences, ccIDs)
}

func sortCardReferences(cr *[]model.CardReference, ccIDs map[string]int) {
	// sorting alphabetically from a-z
	sort.SliceStable(*cr, func(i, j int) bool {
		return (*cr)[i].Card.CardName < (*cr)[j].Card.CardName
	})

	// sorting by card color
	sort.SliceStable(*cr, func(i, j int) bool {
		return ccIDs[(*cr)[i].Card.CardColor] < ccIDs[(*cr)[j].Card.CardColor]
	})
}

// Uses regex to find all direct references to cards (or potentially archetypes) and searches it in the DB.
// If a direct name reference is found in the DB, then it is returned as a suggestion.
func getReferences(ctx context.Context, s string) ([]model.CardReference, []string) {
	namedReferences, referenceOccurrence, archetypalReferences := isolateReferences(ctx, s)

	uniqueReferences := make([]model.CardReference, 0, len(namedReferences))
	for _, card := range namedReferences {
		uniqueReferences = append(uniqueReferences, model.CardReference{Card: card, Occurrences: referenceOccurrence[card.CardID]})
	}

	return uniqueReferences, archetypalReferences
}

func isolateReferences(ctx context.Context, s string) (map[string]model.Card, map[string]int, []string) {
	tokens := quotedStringRegex.FindAllString(s, -1)

	namedReferences, referenceOccurrence, archetypalReferences := buildReferenceObjects(ctx, tokens)

	// get unique archetypes
	uniqueArchetypalReferences := make([]string, 0, len(archetypalReferences))
	for ref := range archetypalReferences {
		uniqueArchetypalReferences = append(uniqueArchetypalReferences, ref)
	}
	sort.Strings(uniqueArchetypalReferences) // needed as source of this array was a map and maps don't have predictable sorting - tests will fail randomly without sort

	return namedReferences, referenceOccurrence, uniqueArchetypalReferences
}

// cycles through tokens - makes DB calls where necessary and attempts to build objects containing direct references (and their occurrences), archetype references
func buildReferenceObjects(ctx context.Context, tokens []string) (map[string]model.Card, map[string]int, map[string]bool) {
	namedReferences := map[string]model.Card{}
	referenceOccurrence := map[string]int{}
	archetypalReferences := map[string]bool{}
	tokenToCardId := map[string]string{} // maps token to its cardID - token will only have cardID if token is found in DB
	totalTokens := len(tokens)

	if totalTokens != 0 {
		for i := 0; i < totalTokens; i++ {
			model.CleanupToken(&tokens[i])
		}

		batchCardData, _ := skcDBInterface.GetDesiredCardsFromDBUsingMultipleCardNames(ctx, tokens)

		for _, token := range tokens {
			// if we already searched the token before we don't need to waste time re-searching it in DB

			// if token is present in archetype slice, skip token
			if _, isPresent := archetypalReferences[token]; isPresent {
				continue
			}

			// if token mapped to a cardId in previous loop, increase number of occurrences by 1 and skip any other processing this iteration as we already did the processing before
			if _, isPresent := tokenToCardId[token]; isPresent {
				referenceOccurrence[tokenToCardId[token]] += 1
				continue
			}

			if card, isPresent := batchCardData.CardInfo[token]; !isPresent {
				// add occurrence of archetype to map
				archetypalReferences[token] = true
			} else {
				// add occurrence of referenced card to maps
				namedReferences[card.CardID] = card
				referenceOccurrence[card.CardID] = 1
				tokenToCardId[token] = card.CardID
			}
		}
	}

	return namedReferences, referenceOccurrence, archetypalReferences
}
